package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"os"
	"strings"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/OZON/stickersFBS"
)

func createCabinetsMarkup(cabinets []db.Cabinet, page int, hasNext bool) models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton
	var row []models.InlineKeyboardButton
	var button models.InlineKeyboardButton
	for _, cabinet := range cabinets {
		row = []models.InlineKeyboardButton{}
		button = models.InlineKeyboardButton{Text: cabinet.Name, CallbackData: fmt.Sprintf("%v%v", CallbackSelectCabinetHandler, cabinet.ID)}
		row = append(row, button)

		keyboard = append(keyboard, row)
	}

	//Добавление кнопок для пагинации
	row = []models.InlineKeyboardButton{}
	if page > 1 {
		button = models.InlineKeyboardButton{Text: "⬅️", CallbackData: CallbackOzonCabinetsHandler + fmt.Sprintf("%v", page-1)}
		row = append(row, button)
	}

	if hasNext {
		button = models.InlineKeyboardButton{Text: "➡️", CallbackData: CallbackOzonCabinetsHandler + fmt.Sprintf("%v", page+1)}
		row = append(row, button)
	}

	if row != nil {
		keyboard = append(keyboard, row)
	}

	//row = []models.InlineKeyboardButton{}
	//button = models.InlineKeyboardButton{Text: "Добавить аккаунт", CallbackData: addParserCallback}
	//row = append(row, button)
	//keyboard = append(keyboard, row)

	row = []models.InlineKeyboardButton{}
	button = models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler}
	row = append(row, button)
	keyboard = append(keyboard, row)

	markup := models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
	return markup
}

func (m *Manager) ozonHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	var cabinets []db.Cabinet
	// Смотрим есть ли артикул в бд
	result := m.db.Where(`"marketplace" = ?`, "ozon").Find(&cabinets)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	text := "Выберите кабинет"
	markup := createCabinetsMarkup(cabinets, 0, false)

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) ozonCabinetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetId := parts[1]

	text := "Кабинет Озон"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: fmt.Sprintf("%v%v", CallbackOzonOrdersHandler, cabinetId)})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Остатки", CallbackData: fmt.Sprintf("%v%v", CallbackOzonStocksHandler, cabinetId)})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: fmt.Sprintf("%v%v", CallbackOzonStickersHandler, cabinetId)})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackOzonHandler})

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

	//parts := strings.Split(update.CallbackQuery.Data, "_")
	//cabinetId := parts[1]

	var cabinets []db.Cabinet

	result := m.db.Where(`"marketplace" = ?`, "ozon").Find(&cabinets)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	titleRange := "!A1"
	fbsRange := "!A2:B1000"
	fboRange := "!D2:E1000"
	returnsRange := "!G2:H1000"

	maxValuesCount, err := OZON.NewService(cabinets[0]).GetOrdersAndReturnsManager().WriteToGoogleSheets(titleRange, fbsRange, fboRange, returnsRange)
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Printf("%v", err)
			return
		}
		return
	}

	maxValuesCount += 3
	titleRange = fmt.Sprintf("!A%v", maxValuesCount)

	maxValuesCount++
	fbsRange = fmt.Sprintf("!A%v:B%v", maxValuesCount, maxValuesCount+1000)
	fboRange = fmt.Sprintf("!D%v:E%v", maxValuesCount, maxValuesCount+1000)
	returnsRange = fmt.Sprintf("!G%v:H%v", maxValuesCount, maxValuesCount+1000)

	_, err = OZON.NewService(cabinets[1]).GetOrdersAndReturnsManager().WriteToGoogleSheets(titleRange, fbsRange, fboRange, returnsRange)
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Printf("%v", err)
			return
		}
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

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetId := parts[1]

	var cabinet db.Cabinet

	result := m.db.Where(`"cabinetsId" = ?`, cabinetId).Find(&cabinet)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	postings := OZON.NewService(cabinet).GetStocksManager().GetPostings()

	stocks := OZON.NewService(cabinet).GetStocksManager().GetStocks()

	filePath, err := generateExcelOzon(postings, stocks, K, "ozon")
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

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetId := parts[1]

	var cabinet db.Cabinet

	result := m.db.Where(`"cabinetsId" = ?`, cabinetId).Find(&cabinet)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	err = OZON.NewService(cabinet).GetStickersFBSManager().GetLabels()
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	filePath := fmt.Sprintf("%v.pdf", stickersFBS.OzonDirectoryPath+"ozon")
	err = sendMediaMessage(ctx, bot, chatId, filePath)
	if err != nil {
		log.Println(err)
		return
	}
	m.ozonService.GetStickersFBSManager().CleanFiles("ozon")

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
