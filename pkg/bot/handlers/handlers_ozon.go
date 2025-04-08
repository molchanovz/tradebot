package bot

import (
	"WildberriesGo_bot/pkg/OZON/StickersFBS"
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"os"
)

func ozonHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	text := "Кабинет Озон"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: CallbackOzonOrdersHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Остатки", CallbackData: CallbackOzonStocksHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: CallbackOzonStickersHandler})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: "START"})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) ozonOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := m.ozonService.GetOrdersAndReturnsManager().WriteToGoogleSheets()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	_, err = sendTextMessage(ctx, bot, chatId, "Заказы озон за вчерашний день были внесены")
	if err != nil {
		log.Printf("%v", err)
		return
	}

}
func (m *Manager) ozonStocksHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {

	K := 1.5

	chatId := update.CallbackQuery.From.ID

	postings := m.ozonService.GetStocksManager().GetPostings()

	stocks := m.ozonService.GetStocksManager().GetStocks()

	filePath, err := generateExcel(postings, stocks, K, "ozon")
	if err != nil {
		log.Println("Ошибка при создании Excel:", err)
		return
	}

	err = sendMediaMessage(ctx, bot, chatId, filePath)
	if err != nil {
		return
	}

	os.Remove(filePath)

}
func (m *Manager) ozonStickersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	text := fmt.Sprintf("Подготовка файла Озон")
	chatId := update.CallbackQuery.From.ID
	message, err := sendTextMessage(ctx, bot, chatId, text)
	if err != nil {
		return
	}

	err = m.ozonService.GetStickersFBSManager().GetLabels()
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	filePath := fmt.Sprintf("%v.pdf", StickersFBS.OzonDirectoryPath+"ozon")
	err = sendMediaMessage(ctx, bot, chatId, filePath)
	if err != nil {
		log.Println(err)
		return
	}
	m.ozonService.GetStickersFBSManager().CleanFiles("ozon")

	text, markup := createStartMarkup()
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
