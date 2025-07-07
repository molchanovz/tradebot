package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"strconv"
	"strings"
	"tradebot/pkg/db"
)

const (
	MessageSettingsHandler     = "/settings"
	CallbackSettingsHandler    = "SETTINGS_"
	CallbackChangeAPIHandler   = "CHANGE-API_"
	CallbackChangeSheetHandler = "CHANGE-SHEET_"
)

func (m *Manager) settingsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	var chatId int64

	if update.Message != nil {
		chatId = update.Message.From.ID
	} else {
		chatId = update.CallbackQuery.From.ID
	}

	text, settingsMarkup := createSettingsMarkup()

	if update.CallbackQuery != nil {
		_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
			MessageID:   update.CallbackQuery.Message.Message.ID,
			ChatID:      chatId,
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
		ChatID:      chatId,
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
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", CallbackData: CallbackSettingsHandler + MarketWb})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", CallbackData: CallbackSettingsHandler + CallbackYandexHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ОЗОН", CallbackData: CallbackSettingsHandler + MarketOzon})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}

	buttonsRow = []models.InlineKeyboardButton{}
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})
	allButtons = append(allButtons, buttonsRow)

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func (m *Manager) selectMpSettingsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("settingsMPHandler неверное кол-во parts")
		return
	}

	mp := parts[1]

	cabinets, err := m.repo.GetCabinets(mp)
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
		ChatID:      chatId,
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
	chatId := update.CallbackQuery.From.ID
	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("settingsMPHandler неверное кол-во parts")
		return
	}

	cabinetId, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Println("ошибка получения cabinetId")
		return
	}

	text, keyboardMarkup := createSettingsMPMarkup(cabinetId)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ChatID:      chatId,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: keyboardMarkup,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}

}

func createSettingsMPMarkup(cabinetId int) (string, models.InlineKeyboardMarkup) {
	startMessage := "Настройки кабинетов. Выбери настройку"
	var buttonsRow []models.InlineKeyboardButton
	var allButtons [][]models.InlineKeyboardButton

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Изменить ключ API", CallbackData: fmt.Sprintf("%v+%v", CallbackChangeAPIHandler, cabinetId)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Изменить таблицу для заказов", CallbackData: fmt.Sprintf("%v+%v", CallbackChangeSheetHandler, cabinetId)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = []models.InlineKeyboardButton{}
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: MessageSettingsHandler})
	allButtons = append(allButtons, buttonsRow)

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func (m *Manager) ChangeApiHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("ChangeApiHandler неверное кол-во parts")
		return
	}

	cabinetId := parts[1]

	user, err := m.repo.GetUserByTgId(chatId)
	if err != nil {
		log.Println("Ошибка получения User")
		return
	}

	if user == nil {
		log.Println("Ошибка получения User")
		return
	}

	user.StatusID = db.WaitingAPI

	err = m.repo.UpdateUser(user)
	if err != nil {
		log.Println("Ошибка обновления User")
		return
	}

	m.ApiMap.Store(chatId, cabinetId)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
		MessageID: update.CallbackQuery.Message.Message.ID,
		ChatID:    chatId,
		Text:      "Отправь новый API ключ",
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}

}

func (m *Manager) ChangeSheetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	parts := strings.Split(update.CallbackQuery.Data, "_")

	if len(parts) != 2 {
		log.Println("ChangeSheetHandler неверное кол-во parts")
		return
	}

	cabinetId := parts[1]

	user, err := m.repo.GetUserByTgId(chatId)
	if err != nil {
		log.Println("Ошибка получения User")
		return
	}

	if user == nil {
		log.Println("Ошибка получения User")
		return
	}

	user.StatusID = db.WaitingSheet

	err = m.repo.UpdateUser(user)
	if err != nil {
		log.Println("Ошибка обновления User")
		return
	}

	m.SheetMap.Store(chatId, cabinetId)

	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID:    nil,
		Text:      "Отправь ссылку на гугл таблицу",
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}

}

func (m *Manager) changeSheet(ctx context.Context, bot *botlib.Bot, chatId int64, message *models.Message) {
	var text string
	if value, ok := m.SheetMap.Load(chatId); ok == true {

		cabinet, err := m.repo.GetCabinetById(value.(string))
		if err != nil {
			log.Println("Ошибка получения кабинета")
			return
		}

		//TODO Добавить поле с гугл таблицей
		//cabinet.Key = message

		err = m.repo.UpdateCabinet(cabinet)
		if err != nil {
			log.Println("Ошибка обновления кабинета")
			return
		}

		m.ApiMap.Delete(chatId)
		text = "Таблица изменена"
	} else {
		text = "Таблица не изменена"
	}

	_, markup := createStartAdminMarkup()

	_, err := bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID:      chatId,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения")
		return
	}

}

func (m *Manager) changeApi(ctx context.Context, bot *botlib.Bot, chatId int64, message *models.Message) {
	var text string
	var cabinet db.Cabinet
	var err error
	if value, ok := m.ApiMap.Load(chatId); ok == true {
		cabinet, err = m.repo.GetCabinetById(value.(string))
		if err != nil {
			log.Println("Ошибка получения кабинета: ", err)
			return
		}

		cabinet.Key = message.Text

		err = m.repo.UpdateCabinet(cabinet)
		if err != nil {
			log.Println("Ошибка обновления кабинета: ", err)
			return
		}

		m.ApiMap.Delete(chatId)
		text = "Ключ изменен"
	} else {
		text = "Ключ не изменен"
	}

	_, err = bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{
		ChatID:    chatId,
		MessageID: message.ID,
	})
	if err != nil {
		log.Println("Ошибка удаления сообщения с API: ", err)
		return
	}

	_, markup := createSettingsMPMarkup(cabinet.ID)

	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
		ChatID:      chatId,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil {
		log.Println("Ошибка отправки сообщения: ", err)
		return
	}

}
