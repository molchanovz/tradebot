package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"tradebot/pkg/client/google"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"

	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	MessageSettingsHandler        = "/settings"
	CallbackSettingsHandler       = "SETTINGS_"
	CallbackChangeAPIHandler      = "CHANGE-API_"
	CallbackChangeSheetHandler    = "CHANGE-SHEET_"
	CallbackShipmentsSheetHandler = "SHIPMENTS-SHEET"

	// shipmentsSheetKey — вместо маркетплейса в CHANGE-SHEET_ и SheetMap: ждём ссылку
	// на таблицу отгрузок, она одна на все кабинеты.
	shipmentsSheetKey = "SHIPMENTS"
)

func (m *Manager) settingsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	var chatID int64

	if update.Message != nil {
		chatID = update.Message.From.ID
	} else {
		chatID = update.CallbackQuery.From.ID
	}

	if m.adminUser(ctx, chatID) == nil {
		if update.Message != nil {
			if _, err := SendTextMessage(ctx, bot, chatID, "Настройки доступны только администраторам бота"); err != nil {
				log.Println(err)
			}
		}
		return
	}

	text, settingsMarkup := createSettingsMarkup()

	if update.CallbackQuery != nil {
		_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
			MessageID:   update.CallbackQuery.Message.Message.ID,
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: settingsMarkup,
		})
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	_, err := bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: settingsMarkup,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func createSettingsMarkup() (string, models.InlineKeyboardMarkup) {
	startMessage := "Настройки кабинетов. Выбери маркетплейс для настройки"
	var buttonsRow []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", CallbackData: CallbackSettingsHandler + CallbackWbHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", CallbackData: CallbackSettingsHandler + CallbackYandexHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ОЗОН", CallbackData: CallbackSettingsHandler + CallbackOzonHandler})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}

	allButtons = append(allButtons, []models.InlineKeyboardButton{{Text: "Таблица отгрузок", CallbackData: CallbackShipmentsSheetHandler}})

	buttonsRow = []models.InlineKeyboardButton{}
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})
	allButtons = append(allButtons, buttonsRow)

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

// selectMpSettingsHandler — настройки маркетплейса: таблица заказов и кабинеты.
func (m *Manager) selectMpSettingsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	user := m.adminUser(ctx, chatID)
	if user == nil {
		return
	}

	m.stopWaitingSheet(ctx, user, chatID)

	mp := strings.TrimPrefix(update.CallbackQuery.Data, CallbackSettingsHandler)

	cabinets, err := m.tm.GetCabinetsByMp(ctx, mp)
	if err != nil {
		log.Println(err)
		return
	}

	text, markup := m.mpSettingsMarkup(mp, cabinets)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

// mpSettingsMarkup — таблица заказов маркетплейса, кнопки для неё и список кабинетов.
func (m *Manager) mpSettingsMarkup(mp string, cabinets tradeplus.Cabinets) (string, models.InlineKeyboardMarkup) {
	callbacks := CallbacksForCabinetMarkup{
		PaginationCallback: CallbackOzonCabinetsHandler,
		SelectCallback:     CallbackSettingsSelectCabinetHandler,
		BackCallback:       MessageSettingsHandler,
	}
	markup := createCabinetsMarkup(cabinets, callbacks, 0, false)

	mk, ok := markets[mp]
	if !ok {
		return "Выберите кабинет", markup
	}

	text := fmt.Sprintf("Настройки %s\n\n%s\n\nКабинеты — нажми, чтобы сменить API-ключ.", mk.name, m.ordersSheetText(mk, cabinets))

	rows := [][]models.InlineKeyboardButton{{{Text: "Сменить таблицу заказов", CallbackData: CallbackChangeSheetHandler + mp}}}
	if cabinets.Enabled().OrdersSpreadsheetID() != "" {
		rows = append(rows, []models.InlineKeyboardButton{{Text: "Занести заказы сейчас", CallbackData: CallbackOrdersSyncHandler + mp}})
	}
	markup.InlineKeyboard = append(rows, markup.InlineKeyboard...)

	return text, markup
}

// ordersSheetText — какая таблица заказов подключена и есть ли к ней доступ.
// Смотрим, как и запись, только включённые кабинеты.
func (m *Manager) ordersSheetText(mk market, cabinets tradeplus.Cabinets) string {
	ids := cabinets.Enabled().OrdersSpreadsheetIDs()
	if len(ids) == 0 {
		return "Таблица заказов не задана."
	}

	text := "Таблица заказов: " + m.sheetLine(ids[0])
	text += fmt.Sprintf("\n\nКаждый день бот смотрит в таблицу и заносит заказы за дни этого месяца по вчера, которых там ещё нет (листы «%s<число>»).", mk.ordersPrefix)

	if len(ids) > 1 {
		text += fmt.Sprintf("\n\nУ кабинетов %s указаны разные таблицы, бот пишет в первую. Пришли ссылку заново, чтобы таблица у всех была одна.", mk.name)
	}

	return text
}

func (m *Manager) settingsMPHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	if m.adminUser(ctx, chatID) == nil {
		return
	}

	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("settingsMPHandler неверное кол-во parts")
		return
	}

	cabinetID, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Println("ошибка получения cabinetID")
		return
	}

	cabinet, err := m.tm.GetCabinetByID(ctx, cabinetID)
	if err != nil {
		log.Println("Ошибка получения кабинета: ", err)
		return
	}

	text, keyboardMarkup := createSettingsMPMarkup(cabinet)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: keyboardMarkup,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}
}

// createSettingsMPMarkup — настройки кабинета. Таблица заказов общая на маркетплейс
// и меняется на экране маркетплейса, куда ведёт «Назад».
func createSettingsMPMarkup(cabinet tradeplus.Cabinet) (string, models.InlineKeyboardMarkup) {
	startMessage := "Настройки кабинета " + cabinet.Name

	back := MessageSettingsHandler
	if cabinet.Marketplace != "" {
		back = CallbackSettingsHandler + cabinet.Marketplace
	}

	allButtons := [][]models.InlineKeyboardButton{
		{{Text: "Изменить ключ API", CallbackData: fmt.Sprintf("%v%v", CallbackChangeAPIHandler, cabinet.ID)}},
		{{Text: "Назад", CallbackData: back}},
	}

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func (m *Manager) ChangeApiHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("ChangeApiHandler неверное кол-во parts")
		return
	}

	cabinetID := parts[1]

	user := m.adminUser(ctx, chatID)
	if user == nil {
		return
	}

	_, err := m.tm.SetUserStatus(ctx, user, db.StatusWaitingAPI)
	if err != nil {
		log.Println("Ошибка обновления статуса User")
		return
	}

	id, err := strconv.Atoi(cabinetID)
	if err != nil {
		log.Println("Ошибка парсинга cabinetId")
		return
	}

	m.APIMap.Store(chatID, id)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID: update.CallbackQuery.Message.Message.ID,
		ChatID:    chatID,
		Text:      "Отправь новый API ключ",
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}
}

// ChangeSheetHandler просит прислать ссылку на таблицу заказов маркетплейса.
func (m *Manager) ChangeSheetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	messageID := update.CallbackQuery.Message.Message.ID

	user := m.adminUser(ctx, chatID)
	if user == nil {
		return
	}

	mp := strings.TrimPrefix(update.CallbackQuery.Data, CallbackChangeSheetHandler)

	var text, cancel string
	if mk, ok := markets[mp]; ok {
		text = fmt.Sprintf("Пришли ссылку на Google-таблицу для заказов %s — она будет общей для всех кабинетов %s.\n\n"+
			"Аккаунту бота нужен доступ на редактирование. Листов «%s<число>», которых нет, бот создаст сам.",
			mk.name, mk.name, mk.ordersPrefix)
		cancel = CallbackSettingsHandler + mp
	} else if mp == shipmentsSheetKey {
		text = "Пришли ссылку на Google-таблицу отгрузок — она будет общей для всех кабинетов.\n\n" +
			"Аккаунту бота нужен доступ на редактирование."
		cancel = CallbackShipmentsSheetHandler
	} else {
		// кнопка из старого меню, где таблица менялась у отдельного кабинета
		_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: messageID, Text: "Эта кнопка устарела — открой /settings заново"})
		if err != nil {
			log.Println("Ошибка отправки сообщения: ", err)
		}
		return
	}

	_, err := m.tm.SetUserStatus(ctx, user, db.StatusWaitingSheet)
	if err != nil {
		log.Println("Ошибка обновления User")
		return
	}

	m.SheetMap.Store(chatID, mp)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID: messageID,
		ChatID:    chatID,
		Text:      text,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: cancel}},
		}},
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}
}

// changeSheet принимает ссылку на таблицу заказов, присланную после «Сменить таблицу заказов».
func (m *Manager) changeSheet(ctx context.Context, bot *botlib.Bot, chatID int64, message *models.Message) {
	value, _ := m.SheetMap.LoadAndDelete(chatID)
	mp, _ := value.(string)

	_, err := bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: message.ID,
	})
	if err != nil {
		log.Println("Ошибка удаления сообщения со ссылкой: ", err)
	}

	var text string
	var markup models.InlineKeyboardMarkup
	if mp == shipmentsSheetKey {
		text, markup = m.setShipmentsSheet(ctx, message.Text)
	} else {
		text, markup = m.setOrdersSheet(ctx, mp, message.Text)
	}

	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения: ", err)
		return
	}
}

// setOrdersSheet проверяет ссылку и делает таблицу таблицей заказов всех кабинетов маркетплейса.
func (m *Manager) setOrdersSheet(ctx context.Context, mp, link string) (string, models.InlineKeyboardMarkup) {
	mk, ok := markets[mp]
	if !ok {
		return "Таблица не изменена: начни заново через /settings", models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Настройки", CallbackData: MessageSettingsHandler}},
		}}
	}

	retry := models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "Прислать другую ссылку", CallbackData: CallbackChangeSheetHandler + mp}},
		{{Text: "Назад", CallbackData: CallbackSettingsHandler + mp}},
	}}

	spreadsheetID, ok := google.ParseSpreadsheetID(link)
	if !ok {
		return "Это не похоже на ссылку на Google-таблицу. Нужна ссылка вида https://docs.google.com/spreadsheets/d/...", retry
	}

	title, sheets, err := m.sheets.Spreadsheet(spreadsheetID)
	if err != nil {
		return fmt.Sprintf("Не получилось открыть таблицу: %v\n\nПроверь, что у аккаунта бота есть доступ на редактирование.", err), retry
	}

	if err = m.tm.SetOrdersSheet(ctx, mp, spreadsheetID); err != nil {
		return fmt.Sprintf("Не получилось сохранить таблицу: %v", err), retry
	}

	text := fmt.Sprintf("Таблица заказов %s: «%s»\n%s", mk.name, title, google.SpreadsheetURL(spreadsheetID))
	if note := daySheetsNote(mk.ordersPrefix, sheets); note != "" {
		text += "\n\n" + note
	}

	return text, models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "Занести заказы сейчас", CallbackData: CallbackOrdersSyncHandler + mp}},
		{{Text: "Назад", CallbackData: CallbackSettingsHandler + mp}},
	}}
}

// shipmentsSheetHandler — таблица отгрузок: она одна на все кабинеты всех маркетплейсов.
func (m *Manager) shipmentsSheetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	user := m.adminUser(ctx, chatID)
	if user == nil {
		return
	}

	m.stopWaitingSheet(ctx, user, chatID)

	cabinets, err := m.tm.GetCabinets(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	text := "Таблица отгрузок (складская) — одна на все кабинеты. Каждую ночь бот заполняет в ней листы «Отправлено на WB / Ozon / Яндекс FBS-<число>» за дни этого месяца по вчера, которые ещё пустые; заполненные листы не трогает.\n\n"
	ids := cabinets.ShipmentsSpreadsheetIDs()
	rows := [][]models.InlineKeyboardButton{{{Text: "Сменить таблицу отгрузок", CallbackData: CallbackChangeSheetHandler + shipmentsSheetKey}}}
	if len(ids) == 0 {
		text += "Таблица не задана."
	} else {
		text += "Сейчас: " + m.sheetLine(ids[0])
		rows = append(rows, []models.InlineKeyboardButton{{Text: "Занести отгрузки сейчас", CallbackData: CallbackShipmentsSyncHandler}})
	}
	if len(ids) > 1 {
		text += "\n\nУ кабинетов указаны разные таблицы отгрузок. Пришли ссылку заново, чтобы таблица у всех была одна."
	}
	rows = append(rows, []models.InlineKeyboardButton{{Text: "Назад", CallbackData: MessageSettingsHandler}})

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: rows},
	})
	if err != nil {
		log.Println(err)
		return
	}
}

// setShipmentsSheet проверяет ссылку и делает таблицу таблицей отгрузок всех кабинетов.
func (m *Manager) setShipmentsSheet(ctx context.Context, link string) (string, models.InlineKeyboardMarkup) {
	back := []models.InlineKeyboardButton{{Text: "Назад", CallbackData: CallbackShipmentsSheetHandler}}
	retry := models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
		{{Text: "Прислать другую ссылку", CallbackData: CallbackChangeSheetHandler + shipmentsSheetKey}},
		back,
	}}

	spreadsheetID, ok := google.ParseSpreadsheetID(link)
	if !ok {
		return "Это не похоже на ссылку на Google-таблицу. Нужна ссылка вида https://docs.google.com/spreadsheets/d/...", retry
	}

	title, _, err := m.sheets.Spreadsheet(spreadsheetID)
	if err != nil {
		return fmt.Sprintf("Не получилось открыть таблицу: %v\n\nПроверь, что у аккаунта бота есть доступ на редактирование.", err), retry
	}

	if err = m.tm.SetShipmentsSheet(ctx, spreadsheetID); err != nil {
		return fmt.Sprintf("Не получилось сохранить таблицу: %v", err), retry
	}

	return fmt.Sprintf("Таблица отгрузок для всех кабинетов: «%s»\n%s", title, google.SpreadsheetURL(spreadsheetID)),
		models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Занести отгрузки сейчас", CallbackData: CallbackShipmentsSyncHandler}},
			back,
		}}
}

// sheetLine — название таблицы и ссылка на неё, либо ссылка и почему таблица не открылась.
func (m *Manager) sheetLine(spreadsheetID string) string {
	title, _, err := m.sheets.Spreadsheet(spreadsheetID)
	if err != nil {
		return fmt.Sprintf("%s\nНет доступа к таблице: %v", google.SpreadsheetURL(spreadsheetID), err)
	}
	return fmt.Sprintf("«%s»\n%s", title, google.SpreadsheetURL(spreadsheetID))
}

// stopWaitingSheet — «Отмена» во время смены таблицы: ссылку больше не ждём.
func (m *Manager) stopWaitingSheet(ctx context.Context, user *tradeplus.User, chatID int64) {
	if user.StatusID != db.StatusWaitingSheet {
		return
	}
	m.SheetMap.Delete(chatID)
	if _, err := m.tm.SetUserStatus(ctx, user, db.StatusEnabled); err != nil {
		log.Println("Ошибка обновления статуса User: ", err)
	}
}

// daySheetsNote подсказывает, каких дневных листов «<prefix><число>» нет в таблице.
func daySheetsNote(prefix string, sheets []google.SheetStat) string {
	// регистр в названиях листов Google не различает
	titles := make(map[string]bool, len(sheets))
	for _, s := range sheets {
		titles[strings.ToLower(s.Title)] = true
	}

	var missing []string
	for i := range 31 {
		if day := strconv.Itoa(i + 1); !titles[strings.ToLower(prefix+day)] {
			missing = append(missing, day)
		}
	}

	switch len(missing) {
	case 0:
		return ""
	case 31:
		return fmt.Sprintf("В таблице нет ни одного листа «%s<число>» — точно та таблица? Листы бот создаст сам.", prefix)
	default:
		return fmt.Sprintf("Нет листов «%s<число>» за числа %s — бот создаст их сам.", prefix, strings.Join(missing, ", "))
	}
}

func (m *Manager) changeAPI(ctx context.Context, bot *botlib.Bot, chatID int64, message *models.Message) {
	var text string
	var cabinet tradeplus.Cabinet
	if value, ok := m.APIMap.Load(chatID); ok {
		var err error
		cabinet, err = m.tm.GetCabinetByID(ctx, value.(int))
		if err != nil {
			log.Println("Ошибка получения кабинета")
			return
		}

		cabinet.Key = message.Text

		err = m.tm.UpdateCabinet(ctx, cabinet)
		if err != nil {
			log.Println("Ошибка обновления кабинета")
			return
		}

		m.APIMap.Delete(chatID)
		text = "Ключ изменен"
	} else {
		text = "Ключ не изменен"
	}

	_, err := bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: message.ID,
	})
	if err != nil {
		log.Println("Ошибка удаления сообщения с API: ", err)
		return
	}

	_, markup := createSettingsMPMarkup(cabinet)

	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения: ", err)
		return
	}
}

// adminUser возвращает пользователя, если он администратор: настройки и запись
// в таблицы доступны только администраторам.
func (m *Manager) adminUser(ctx context.Context, chatID int64) *tradeplus.User {
	user, err := m.tm.UserByChatID(ctx, chatID)
	if err != nil {
		log.Println("Ошибка получения User: ", err)
		return nil
	}
	if user == nil || !user.IsAdmin {
		return nil
	}
	return user
}
