package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/WB/stickersFbs"
	"tradebot/pkg/marketplaces/YANDEX/yandex_stickers_fbs"
)

func (m *Manager) yandexHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	text := "Кабинет Яндекс"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: CallbackYandexFbsHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: CallbackYandexOrdersHandler})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: "START"})

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

	err := m.db.Model(&db.User{}).Where(`"tgId" = ?`, chatId).Updates(db.User{
		TgId:     chatId,
		StatusId: db.WaitingYaState,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления WaitingYaState пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен WaitingYaState", chatId)

	text := fmt.Sprintf("Отправь мне номер отгрузки")
	var buttonBack []models.InlineKeyboardButton

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: "START"})

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

	_, err = sendTextMessage(ctx, bot, chatId, "Заказы яндекс за вчерашний день были внесены")
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) getYandexFbsDEPRECATED(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	text := fmt.Sprintf("Подготовка файла Яндекс")
	message, err := sendTextMessage(ctx, bot, chatId, text)
	if err != nil {
		return
	}

	_, err = m.yandexService.GetStickersFbsManager().GetOrdersInfo(supplyId, nil)
	if err != nil {
		_, err := sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		filePath := fmt.Sprintf("%v.pdf", yandex_stickers_fbs.YaDirectoryPath+supplyId)
		sendMediaMessage(ctx, bot, chatId, filePath)
		yandex_stickers_fbs.CleanFiles(supplyId)
	}

	text, markup := createStartAdminMarkup()
	_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

	_, err = bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{ChatID: chatId, MessageID: message.ID})
	if err != nil {
		return
	}

}
func (m *Manager) getYandexFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {

	done := make(chan string)
	progressChan := make(chan stickersFbs.Progress)

	go func() {
		filePath, err := m.yandexService.GetStickersFbsManager().GetOrdersInfo(supplyId, progressChan)
		if err != nil {
			log.Println("Ошибка при получении файла:", err)
			done <- ""
			return
		}
		done <- filePath
	}()

	m.WaitReadyFile(ctx, bot, chatId, progressChan, done)

}
