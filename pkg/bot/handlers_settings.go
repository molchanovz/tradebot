package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"

	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	MessageSettingsHandler     = "/settings"
	CallbackSettingsHandler    = "SETTINGS_"
	CallbackChangeAPIHandler   = "CHANGE-API_"
	CallbackChangeSheetHandler = "CHANGE-SHEET_"
)

func (m *Manager) settingsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	var chatID int64

	if update.Message != nil {
		chatID = update.Message.From.ID
	} else {
		chatID = update.CallbackQuery.From.ID
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

	buttonsRow = []models.InlineKeyboardButton{}
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})
	allButtons = append(allButtons, buttonsRow)

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func (m *Manager) selectMpSettingsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("settingsMPHandler неверное кол-во parts")
		return
	}

	mp := parts[1]

	cabinets, err := m.bl.GetCabinetsByMp(ctx, mp)
	if err != nil {
		log.Println(err)
		return
	}

	callbacks := CallbacksForCabinetMarkup{
		PaginationCallback: CallbackOzonCabinetsHandler,
		SelectCallback:     CallbackSettingsSelectCabinetHandler,
		BackCallback:       MessageSettingsHandler,
	}

	markup := createCabinetsMarkup(cabinets, callbacks, 0, false)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ChatID:      chatID,
		Text:        "Выберите кабинет",
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: markup,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func (m *Manager) settingsMPHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
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

	text, keyboardMarkup := createSettingsMPMarkup(cabinetID)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: keyboardMarkup,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}
}

func createSettingsMPMarkup(cabinetID int) (string, models.InlineKeyboardMarkup) {
	startMessage := "Настройки кабинетов. Выбери настройку"
	var buttonsRow []models.InlineKeyboardButton
	var allButtons [][]models.InlineKeyboardButton

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Изменить ключ API", CallbackData: fmt.Sprintf("%v+%v", CallbackChangeAPIHandler, cabinetID)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Изменить таблицу для заказов", CallbackData: fmt.Sprintf("%v+%v", CallbackChangeSheetHandler, cabinetID)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: MessageSettingsHandler})
	allButtons = append(allButtons, buttonsRow)

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

	user, err := m.bl.UserByChatID(ctx, chatID)
	if err != nil {
		log.Println("Ошибка получения User")
		return
	}

	if user == nil {
		log.Println("Ошибка получения User")
		return
	}

	_, err = m.bl.SetUserStatus(ctx, user, db.StatusWaitingAPI)
	if err != nil {
		log.Println("Ошибка обновления статуса User")
		return
	}

	m.APIMap.Store(chatID, cabinetID)

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

func (m *Manager) ChangeSheetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("ChangeSheetHandler неверное кол-во parts")
		return
	}

	cabinetID := parts[1]

	user, err := m.bl.UserByChatID(ctx, chatID)
	if err != nil {
		log.Println("Ошибка получения User")
		return
	}

	if user == nil {
		log.Println("Ошибка получения User")
		return
	}

	_, err = m.bl.SetUserStatus(ctx, user, db.StatusWaitingSheet)
	if err != nil {
		log.Println("Ошибка обновления User")
		return
	}

	m.SheetMap.Store(chatID, cabinetID)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID: update.CallbackQuery.Message.Message.ID,
		ChatID:    chatID,
		Text:      "Отправь ссылку на гугл таблицу",
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}
}

func (m *Manager) changeSheet(ctx context.Context, bot *botlib.Bot, chatID int64, message *models.Message) {
	var text string
	var cabinet tradeplus.Cabinet

	if value, ok := m.SheetMap.Load(chatID); ok {
		var err error
		cabinet, err = m.bl.GetCabinetByID(ctx, value.(int))
		if err != nil {
			log.Println("Ошибка получения кабинета")
			return
		}

		cabinet.SheetLink = &message.Text

		err = m.bl.UpdateCabinet(ctx, cabinet)
		if err != nil {
			log.Println("Ошибка обновления кабинета")
			return
		}

		m.APIMap.Delete(chatID)
		text = "Таблица изменена"
	} else {
		text = "Таблица не изменена"
	}

	_, err := bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: message.ID,
	})
	if err != nil {
		log.Println("Ошибка удаления сообщения с API: ", err)
		return
	}

	_, markup := createSettingsMPMarkup(cabinet.ID)

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

func (m *Manager) changeAPI(ctx context.Context, bot *botlib.Bot, chatID int64, message *models.Message) {
	var text string
	var cabinet tradeplus.Cabinet
	if value, ok := m.APIMap.Load(chatID); ok {
		var err error
		cabinet, err = m.bl.GetCabinetByID(ctx, value.(int))
		if err != nil {
			log.Println("Ошибка получения кабинета")
			return
		}

		cabinet.Key = message.Text

		err = m.bl.UpdateCabinet(ctx, cabinet)
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

	_, markup := createSettingsMPMarkup(cabinet.ID)

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
