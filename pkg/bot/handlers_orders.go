package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
	"tradebot/pkg/tradeplus/wb"
	"tradebot/pkg/tradeplus/yandex"

	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	CallbackOrdersSyncHandler    = "ORDERS-SYNC_"
	CallbackShipmentsSyncHandler = "SHIPMENTS-SYNC"
)

// errNoShipmentsTable — у кабинетов маркетплейса не задана таблица отгрузок.
var errNoShipmentsTable = errors.New("не задана таблица отгрузок")

// market — маркетплейс с ежедневными таблицами: заказов (sheetLink) и отгрузок
// (shipmentsSheetId, складская таблица).
type market struct {
	name         string
	ordersPrefix string
	orders       func(tradeplus.Cabinets) (tradeplus.OrdersWriter, error)
	shipments    func(context.Context, tradeplus.Cabinets) ([]tradeplus.ShipmentsWriter, error)
	// Каждая таблица маркетплейса пишется в один поток: cron и кнопка в боте не мешают друг другу.
	ordersMu    *sync.Mutex
	shipmentsMu *sync.Mutex
}

var markets = map[string]market{
	db.MarketWB: {
		name: "WB", ordersPrefix: wb.OrdersSheetPrefix, orders: wb.NewOrdersWriter, shipments: wb.NewShipmentsWriters,
		ordersMu: new(sync.Mutex), shipmentsMu: new(sync.Mutex),
	},
	db.MarketOzon: {
		name: "Ozon", ordersPrefix: ozon.OrdersSheetPrefix, orders: ozon.NewOrdersWriter, shipments: ozon.NewShipmentsWriters,
		ordersMu: new(sync.Mutex), shipmentsMu: new(sync.Mutex),
	},
	db.MarketYandex: {
		name: "Яндекс Маркет", ordersPrefix: yandex.OrdersSheetPrefix, orders: yandex.NewOrdersWriter, shipments: yandex.NewShipmentsWriters,
		ordersMu: new(sync.Mutex), shipmentsMu: new(sync.Mutex),
	},
}

// marketsOrder — маркетплейсы в том порядке, в каком бот о них пишет.
var marketsOrder = []string{db.MarketWB, db.MarketOzon, db.MarketYandex}

// SyncOrders заносит в таблицу заказов маркетплейса дни этого месяца по вчера,
// которых в ней ещё нет (см. tradeplus.OrdersDays). Если по маркетплейсу уже идёт
// запись — например, запущенная кнопкой в боте, — ждёт её окончания.
func (m *Manager) SyncOrders(ctx context.Context, mp string) ([]time.Time, error) {
	mk, ok := markets[mp]
	if !ok {
		return nil, fmt.Errorf("неизвестный маркетплейс %q", mp)
	}

	mk.ordersMu.Lock()
	defer mk.ordersMu.Unlock()
	return m.syncOrders(ctx, mk, mp, nil)
}

// SyncShipments заполняет в складской таблице листы отгрузок маркетплейса за дни
// этого месяца по вчера, которых там нет или которые пустые (см.
// tradeplus.ShipmentsWriter). Если таблица отгрузок не задана, ничего не делает.
// Если по маркетплейсу уже идёт запись, ждёт её окончания.
func (m *Manager) SyncShipments(ctx context.Context, mp string) ([]time.Time, error) {
	mk, ok := markets[mp]
	if !ok {
		return nil, fmt.Errorf("неизвестный маркетплейс %q", mp)
	}

	mk.shipmentsMu.Lock()
	defer mk.shipmentsMu.Unlock()
	written, err := m.syncShipments(ctx, mk, mp, nil)
	if errors.Is(err, errNoShipmentsTable) {
		return nil, nil
	}
	return written, err
}

// syncOrders — запись без блокировки; progress — см. tradeplus.OrdersWriter.Progress.
func (m *Manager) syncOrders(ctx context.Context, mk market, mp string, progress func(written []time.Time, total int)) ([]time.Time, error) {
	cabinets, err := m.tm.GetCabinetsByMp(ctx, mp)
	if err != nil {
		return nil, fmt.Errorf("fetch cabinets failed: %w", err)
	}

	w, err := mk.orders(tradeplus.Cabinets(cabinets).Enabled())
	if err != nil {
		return nil, err
	}
	w.Progress = progress
	return w.Sync(ctx, time.Now())
}

// syncShipments — запись без блокировки; progress — см. tradeplus.ShipmentsWriter.Progress.
func (m *Manager) syncShipments(ctx context.Context, mk market, mp string, progress func(written []time.Time, total int)) ([]time.Time, error) {
	cabinets, err := m.tm.GetCabinetsByMp(ctx, mp)
	if err != nil {
		return nil, fmt.Errorf("fetch cabinets failed: %w", err)
	}

	writers, err := mk.shipments(ctx, tradeplus.Cabinets(cabinets).Enabled())
	if err != nil {
		return nil, err
	}
	if len(writers) == 0 {
		return nil, errNoShipmentsTable
	}

	var written []time.Time
	var errs []error
	for _, w := range writers {
		w.Progress = progress
		days, err := w.Sync(ctx, time.Now())
		written = append(written, days...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return written, errors.Join(errs...)
}

// ordersSyncHandler — кнопка «Занести заказы сейчас»: то же, что делает cron, но сразу.
func (m *Manager) ordersSyncHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	_, err := bot.AnswerCallbackQuery(ctx, &botlib.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})
	if err != nil {
		log.Printf("%v", err)
	}

	if m.adminUser(ctx, chatID) == nil {
		return
	}

	mp := strings.TrimPrefix(update.CallbackQuery.Data, CallbackOrdersSyncHandler)
	mk, ok := markets[mp]
	if !ok {
		return
	}

	if !mk.ordersMu.TryLock() {
		if _, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Заказы %s уже заносятся — дождись окончания", mk.name)); err != nil {
			log.Printf("%v", err)
		}
		return
	}
	defer mk.ordersMu.Unlock()

	startMsg, err := SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Смотрю, каких дней нет в таблице заказов %s...", mk.name))
	if err != nil {
		log.Printf("%v", err)
	}

	written, syncErr := m.syncOrders(ctx, mk, mp, func(written []time.Time, total int) {
		editProgress(ctx, bot, chatID, startMsg, total, ordersProgressText(mk.name, mp, len(written), total))
	})

	deleteMessage(ctx, bot, chatID, startMsg)

	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID: chatID,
		Text:   ordersSyncText(mk.name, written, syncErr),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: CallbackSettingsHandler + mp}},
		}},
	})
	if err != nil {
		log.Printf("%v", err)
	}
}

// shipmentsSyncHandler — кнопка «Занести отгрузки сейчас»: то же, что делают
// cron-задачи отгрузок, сразу по всем маркетплейсам.
func (m *Manager) shipmentsSyncHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	_, err := bot.AnswerCallbackQuery(ctx, &botlib.AnswerCallbackQueryParams{CallbackQueryID: update.CallbackQuery.ID})
	if err != nil {
		log.Printf("%v", err)
	}

	if m.adminUser(ctx, chatID) == nil {
		return
	}

	startMsg, err := SendTextMessage(ctx, bot, chatID, "Смотрю, каких дней нет в складской таблице...")
	if err != nil {
		log.Printf("%v", err)
	}

	lines := make([]string, 0, len(marketsOrder))
	for _, mp := range marketsOrder {
		mk := markets[mp]
		if !mk.shipmentsMu.TryLock() {
			lines = append(lines, fmt.Sprintf("%s: отгрузки уже заносятся", mk.name))
			continue
		}
		lines = append(lines, func() string {
			defer mk.shipmentsMu.Unlock()
			written, syncErr := m.syncShipments(ctx, mk, mp, func(written []time.Time, total int) {
				editProgress(ctx, bot, chatID, startMsg, total, fmt.Sprintf("Заношу отгрузки %s: внесено %d из %d дн.", mk.name, len(written), total))
			})
			return shipmentsSyncLine(mk.name, written, syncErr)
		}())
	}

	deleteMessage(ctx, bot, chatID, startMsg)

	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID: chatID,
		Text:   strings.Join(lines, "\n\n"),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: CallbackShipmentsSheetHandler}},
		}},
	})
	if err != nil {
		log.Printf("%v", err)
	}
}

// ordersProgressText — сколько дней уже внесено, пока идёт запись.
func ordersProgressText(name, mp string, done, total int) string {
	text := fmt.Sprintf("Заношу заказы %s: внесено %d из %d дн.", name, done, total)
	if mp == db.MarketWB && done < total {
		// у WB между днями пауза около минуты (статистика — запрос в минуту)
		text += fmt.Sprintf("\nWB отдаёт статистику раз в минуту, осталось примерно %d мин. Дни появляются в таблице по мере загрузки.", total-done)
	}
	return text
}

// ordersSyncText — итог записи для пользователя.
func ordersSyncText(name string, written []time.Time, err error) string {
	if len(written) == 0 && err == nil {
		return fmt.Sprintf("Пропусков нет: все дни этого месяца по вчера уже в таблице заказов %s", name)
	}

	var parts []string
	if len(written) > 0 {
		parts = append(parts, fmt.Sprintf("Заказы %s внесены за %s", name, tradeplus.FormatDays(written)))
	}
	if err != nil {
		parts = append(parts, fmt.Sprintf("Не получилось: %v", err))
	}
	return strings.Join(parts, "\n\n")
}

// shipmentsSyncLine — итог записи отгрузок одного маркетплейса.
func shipmentsSyncLine(name string, written []time.Time, err error) string {
	if errors.Is(err, errNoShipmentsTable) {
		return fmt.Sprintf("%s: таблица отгрузок не задана", name)
	}

	var parts []string
	if len(written) > 0 {
		parts = append(parts, fmt.Sprintf("%s: внесено за %s", name, tradeplus.FormatDays(written)))
	}
	if err != nil {
		parts = append(parts, fmt.Sprintf("%s: не получилось: %v", name, err))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%s: пропусков нет", name)
	}
	return strings.Join(parts, "\n")
}

// editProgress показывает ход записи в сообщении msg.
func editProgress(ctx context.Context, bot *botlib.Bot, chatID int64, msg *models.Message, total int, text string) {
	if msg == nil || total == 0 {
		return
	}
	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: msg.ID, Text: text})
	if err != nil {
		log.Printf("%v", err)
	}
}

func deleteMessage(ctx context.Context, bot *botlib.Bot, chatID int64, msg *models.Message) {
	if msg == nil {
		return
	}
	if _, err := bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{ChatID: chatID, MessageID: msg.ID}); err != nil {
		log.Printf("%v", err)
	}
}
