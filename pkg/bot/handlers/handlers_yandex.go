package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"tradebot/pkg/db"
	"tradebot/pkg/fbsPrinter"
	"tradebot/pkg/marketplaces/YANDEX"
)

const (
	CallbackYandexHandler       = "YANDEX"
	CallbackYandexFbsHandler    = "YANDEX-FBS"
	CallbackYandexOrdersHandler = "YANDEX-ORDERS"
)

func (m *Manager) yandexHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	text := "Кабинет Яндекс"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: CallbackYandexFbsHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: CallbackYandexOrdersHandler})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) yandexFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	user, err := m.repo.GetUserByTgId(chatId)
	if err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		return
	}

	user.StatusID = db.WaitingYaState
	err = m.repo.UpdateUser(user)
	if err != nil {
		log.Println("Ошибка обновления WaitingYaState пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен WaitingYaState", chatId)

	text := fmt.Sprintf("Отправь мне номер отгрузки")
	var buttonBack []models.InlineKeyboardButton

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{MessageID: update.CallbackQuery.Message.Message.ID, ChatID: chatId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) yandexOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := m.yandexService.GetOrdersAndReturnsManager().WriteToGoogleSheets()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	_, err = SendTextMessage(ctx, bot, chatId, "Заказы яндекс за вчерашний день были внесены")
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) getYandexFbsDEPRECATED(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	text := fmt.Sprintf("Подготовка файла Яндекс")
	message, err := SendTextMessage(ctx, bot, chatId, text)
	if err != nil {
		return
	}

	_, err = m.yandexService.GetStickersFbsManager().GetOrdersInfo(supplyId, nil)
	if err != nil {
		_, err := SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		filePath := fmt.Sprintf("%v.pdf", YANDEX.YaDirectoryPath+supplyId)
		SendMediaMessage(ctx, bot, chatId, filePath)
		YANDEX.CleanFiles(supplyId)
	}

	text, markup := createStartAdminMarkup()
	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Println(fmt.Sprintf("ошибка отправки сообщения %v", err))
		return
	}

	_, err = bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{ChatID: chatId, MessageID: message.ID})
	if err != nil {
		return
	}

}
func (m *Manager) getYandexFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {

	done := make(chan []string)
	progressChan := make(chan fbsPrinter.Progress)
	errChan := make(chan error)
	var filePaths []string

	go func() {
		filePath, err := m.yandexService.GetStickersFbsManager().GetOrdersInfo(supplyId, progressChan)
		if err != nil {
			log.Println("Ошибка при получении файла:", err)
			errChan <- err
			return
		}

		filePaths = append(filePaths, filePath)

		done <- filePaths
	}()

	err := WaitReadyFile(ctx, bot, chatId, progressChan, done, errChan)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			return
		}
		return
	}
}
