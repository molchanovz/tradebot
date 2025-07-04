package bot

import (
	"context"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"strings"
)

const (
	MessageSettingsHandler  = "/settings"
	CallbackSettingsHandler = "SETTINGS_"
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

	cabinetId := parts[1]

	text, keyboardMarkup := createSettingsMPMarkup(cabinetId)

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
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

func createSettingsMPMarkup(cabinetId string) (string, models.InlineKeyboardMarkup) {
	startMessage := "Настройки кабинетов. Выбери настройку"
	var buttonsRow []models.InlineKeyboardButton
	var allButtons [][]models.InlineKeyboardButton

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Изменить ключ API", CallbackData: CallbackSettingsHandler + cabinetId})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Изменить таблицу для заказов", CallbackData: CallbackSettingsHandler + cabinetId})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = []models.InlineKeyboardButton{}
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: MessageSettingsHandler})
	allButtons = append(allButtons, buttonsRow)

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}
