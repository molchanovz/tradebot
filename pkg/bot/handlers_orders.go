package bot

import (
	"context"
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

const CallbackOrdersSyncHandler = "ORDERS-SYNC_"

// ordersMarket — маркетплейс с ежедневной таблицей заказов (sheetLink).
type ordersMarket struct {
	name   string
	prefix string
	writer func(tradeplus.Cabinets) (tradeplus.OrdersWriter, error)
	// mu — таблица заказов маркетплейса пишется в один поток: cron и кнопка в боте не мешают друг другу.
	mu *sync.Mutex
}

var ordersMarkets = map[string]ordersMarket{
	db.MarketWB:     {name: "WB", prefix: wb.OrdersSheetPrefix, writer: wb.NewOrdersWriter, mu: new(sync.Mutex)},
	db.MarketOzon:   {name: "Ozon", prefix: ozon.OrdersSheetPrefix, writer: ozon.NewOrdersWriter, mu: new(sync.Mutex)},
	db.MarketYandex: {name: "Яндекс Маркет", prefix: yandex.OrdersSheetPrefix, writer: yandex.NewOrdersWriter, mu: new(sync.Mutex)},
}

// SyncOrders заносит в таблицу заказов маркетплейса дни этого месяца по вчера,
// которых в ней ещё нет (см. tradeplus.OrdersDays). Если по маркетплейсу уже идёт
// запись — например, запущенная кнопкой в боте, — ждёт её окончания.
func (m *Manager) SyncOrders(ctx context.Context, mp string) ([]time.Time, error) {
	market, ok := ordersMarkets[mp]
	if !ok {
		return nil, fmt.Errorf("неизвестный маркетплейс %q", mp)
	}

	market.mu.Lock()
	defer market.mu.Unlock()
	return m.syncOrders(ctx, market, mp)
}

func (m *Manager) syncOrders(ctx context.Context, market ordersMarket, mp string) ([]time.Time, error) {
	cabinets, err := m.tm.GetCabinetsByMp(ctx, mp)
	if err != nil {
		return nil, fmt.Errorf("fetch cabinets failed: %w", err)
	}

	w, err := market.writer(tradeplus.Cabinets(cabinets).OrdersCabinets())
	if err != nil {
		return nil, err
	}
	return w.Sync(ctx, time.Now())
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
	market, ok := ordersMarkets[mp]
	if !ok {
		return
	}

	if !market.mu.TryLock() {
		if _, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Заказы %s уже заносятся — дождись окончания", market.name)); err != nil {
			log.Printf("%v", err)
		}
		return
	}
	defer market.mu.Unlock()

	text := fmt.Sprintf("Заношу заказы %s в таблицу...", market.name)
	if mp == db.MarketWB {
		text += "\nЕсли пропущено несколько дней, на каждый уйдёт около минуты: WB отдаёт статистику раз в минуту."
	}
	startMsg, err := SendTextMessage(ctx, bot, chatID, text)
	if err != nil {
		log.Printf("%v", err)
	}

	written, syncErr := m.syncOrders(ctx, market, mp)

	if startMsg != nil {
		if _, err = bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{ChatID: chatID, MessageID: startMsg.ID}); err != nil {
			log.Printf("%v", err)
		}
	}

	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID: chatID,
		Text:   ordersSyncText(market.name, written, syncErr),
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: CallbackSettingsHandler + mp}},
		}},
	})
	if err != nil {
		log.Printf("%v", err)
	}
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
