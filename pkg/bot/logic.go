package bot

import (
	"WildberriesGo_bot/pkg/OZON/ozon_orders_returns"
	"WildberriesGo_bot/pkg/OZON/ozon_stocks"
	"WildberriesGo_bot/pkg/WB/wb_orders_returns"
	"WildberriesGo_bot/pkg/WB/wb_stickers_fbs"
	"WildberriesGo_bot/pkg/WB/wb_stocks_analyze"
	"WildberriesGo_bot/pkg/YANDEX/yandex_orders_returns"
	"WildberriesGo_bot/pkg/YANDEX/yandex_stickers_fbs"
	"WildberriesGo_bot/pkg/api/wb"
	"WildberriesGo_bot/pkg/db"
	"context"
	"errors"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	CallbackWbHandler           = "WB"
	CallbackYandexHandler       = "YANDEX"
	CallbackOzonHandler         = "OZON"
	CallbackWbFbsHandler        = "WB_FBS"
	CallbackYandexFbsHandler    = "YANDEX_FBS"
	CallbackWbOrdersHandler     = "WB_ORDERS"
	CallbackYandexOrdersHandler = "YANDEX_ORDERS"
	CallbackOzonOrdersHandler   = "OZON_ORDERS"
	CallbackOzonStocksHandler   = "OZON_STOCKS"
	CallbackWbStocksHandler     = "WB_STOCKS"
)

type Manager struct {
	b        *botlib.Bot
	db       *gorm.DB
	myChatId string
}

func NewBotManager(db *gorm.DB, myChatId string) *Manager {
	return &Manager{
		db:       db,
		myChatId: myChatId,
	}
}

func (m *Manager) SetBot(bot *botlib.Bot) {
	m.b = bot
}
func (m *Manager) GetBot() *botlib.Bot {
	return m.b
}

func (m *Manager) RegisterBotHandlers() {
	m.b.RegisterHandler(botlib.HandlerTypeMessageText, "/start", botlib.MatchTypePrefix, m.startHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "START", botlib.MatchTypePrefix, m.startHandler)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbHandler, botlib.MatchTypeExact, wbHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexHandler, botlib.MatchTypeExact, yandexHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonHandler, botlib.MatchTypeExact, ozonHandler)

	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbFbsHandler, botlib.MatchTypeExact, m.wbFbsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexFbsHandler, botlib.MatchTypeExact, m.yandexFbsHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbOrdersHandler, botlib.MatchTypePrefix, wbOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexOrdersHandler, botlib.MatchTypePrefix, yandexOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonOrdersHandler, botlib.MatchTypePrefix, ozonOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStocksHandler, botlib.MatchTypePrefix, ozonStocksHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbStocksHandler, botlib.MatchTypePrefix, wbStocksHandler)

	//b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "YANDEX_FBS", botlib.MatchTypePrefix, wbOrdersHandler)

}

func (m *Manager) DefaultHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.Message.From.ID
	message := update.Message.Text

	var user db.User
	// Смотрим есть ли артикул в бд
	result := m.db.Where("chatId = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding stocksApi:", result.Error)
	}

	switch user.State {
	case db.DefaultState:
		{
			sendTextMessage(ctx, bot, chatId, "Не понял тебя. Нажми /start еще раз")
		}
	case db.WaitingWbState:
		{
			getWbFbs(ctx, bot, chatId, message)
		}
	case db.WaitingYaState:
		{
			getYandexFbs(ctx, bot, chatId, message)
		}
	}

	err := m.db.Model(&db.User{}).Where("chatId = ?", chatId).Updates(db.User{
		ChatId: chatId,
		State:  db.DefaultState,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления DefaultState пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен DefaultState", chatId)

}

func createStartMarkup() (string, models.InlineKeyboardMarkup) {
	startMessage := "Выбери маркетплейс для работы"
	var buttonsRow []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", CallbackData: CallbackWbHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", CallbackData: CallbackYandexHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ОЗОН", CallbackData: CallbackOzonHandler})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func (m *Manager) startHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	var chatId int64
	var text string

	startMessage, markup := createStartMarkup()

	if update.Message != nil {
		chatId = update.Message.From.ID
		name := update.Message.From.FirstName
		text = fmt.Sprintf("Привет, %v. %v", name, startMessage)

		_, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text, ReplyMarkup: markup})
		if err != nil {
			log.Printf("%v", err)
			return
		}

	} else {
		chatId = update.CallbackQuery.From.ID
		messageId := update.CallbackQuery.Message.Message.ID
		text = startMessage

		_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{MessageID: messageId, ChatID: chatId, Text: text, ReplyMarkup: markup})
		if err != nil {
			log.Printf("%v", err)
			return
		}
	}

	var user db.User
	// Смотрим есть ли юзер в бд
	result := m.db.Where("chatId = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding chatId: ", result.Error)
	}

	// если юзера нет - заполняем бд
	if user.ChatId == 0 {
		user := db.User{ChatId: chatId, State: db.DefaultState}
		err := m.db.Create(&user).Error
		if err != nil {
			log.Println("Ошибка создания пользователя: ", err)
		}
		log.Printf("Пользователь %v создан", chatId)
	} else {
		err := m.db.Model(&db.User{}).Where("chatId = ?", chatId).Updates(db.User{
			ChatId: chatId,
			State:  db.DefaultState,
		}).Error
		if err != nil {
			log.Println("Ошибка обновления DefaultState пользователя: ", err)
		}
		log.Printf("У пользователя %v обновлен DefaultState", chatId)
	}
}

func wbHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	text := "Кабинет ВБ"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: CallbackWbFbsHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: CallbackWbOrdersHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Остатки", CallbackData: CallbackWbStocksHandler})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: "START"})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}
func yandexHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
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
func ozonHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	text := "Кабинет Озон"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: CallbackOzonOrdersHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Остатки", CallbackData: CallbackOzonStocksHandler})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: "START"})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) wbFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := m.db.Model(&db.User{}).Where("chatId = ?", chatId).Updates(db.User{
		ChatId: chatId,
		State:  db.WaitingWbState,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления WaitingWbState пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен WaitingWbState", chatId)

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
func (m *Manager) yandexFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := m.db.Model(&db.User{}).Where("chatId = ?", chatId).Updates(db.User{
		ChatId: chatId,
		State:  db.WaitingYaState,
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

func getWbFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	wildberriesKey, err := initEnv("variables.env", "API_KEY_WB")
	if err != nil {
		return
	}

	text := fmt.Sprintf("Подготовка файла ВБ")
	message, err := sendTextMessage(ctx, bot, chatId, text)
	if err != nil {
		return
	}

	err = wb_stickers_fbs.GetReadyFile(wildberriesKey, supplyId)
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	filePath := fmt.Sprintf("%v%v.pdf", wb_stickers_fbs.DirectoryPath, supplyId)
	err = sendMediaMessage(ctx, bot, chatId, filePath)
	if err != nil {
		log.Println(err)
		return
	}
	wb_stickers_fbs.Clean_files(supplyId)

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
func wbOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	wildberriesKey, err := initEnv("variables.env", "API_KEY_WB")
	if err != nil {
		log.Println(err)
		return
	}

	err = wb_orders_returns.WriteToGoogleSheets(wildberriesKey)

	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		_, err = sendTextMessage(ctx, bot, chatId, "Заказы вб за вчерашний день были внесены")
		if err != nil {
			log.Println(err)
			return
		}
	}
}
func yandexOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	yandexToken, err := initEnv("variables.env", "yandexToken")
	if err != nil {
		log.Panic(err)
	}
	err = yandex_orders_returns.WriteToGoogleSheets(yandexToken)
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
func ozonOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	ClientId, err := initEnv("variables.env", "ClientId")
	if err != nil {
		log.Panic(err)
	}
	OzonKey, err := initEnv("variables.env", "OzonKey")
	if err != nil {
		log.Panic(err)
	}
	ozon_orders_returns.WriteToGoogleSheets(ClientId, OzonKey)

	_, err = sendTextMessage(ctx, bot, chatId, "Заказы озон за вчерашний день были внесены")
	if err != nil {
		log.Printf("%v", err)
		return
	}

}
func ozonStocksHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	daysAgo := 14
	K := 1.5

	chatId := update.CallbackQuery.From.ID
	ClientId, err := initEnv("variables.env", "ClientId")
	if err != nil {
		log.Println(err)
		return
	}
	OzonKey, err := initEnv("variables.env", "OzonKey")
	if err != nil {
		log.Println(err)
		return
	}

	postings := ozon_stocks.GetPostings(ClientId, OzonKey, daysAgo)

	stocks := ozon_stocks.GetStocks(ClientId, OzonKey)

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
func wbStocksHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	daysAgo := 14
	K := 100.0

	chatId := update.CallbackQuery.From.ID

	WbKey, err := initEnv("variables.env", "API_KEY_WB")
	if err != nil {
		log.Println(err)
		return
	}

	orders := wb_stocks_analyze.GetOrders(WbKey, daysAgo)
	str := strings.Builder{}
	for cluster, value := range orders {
		if cluster == "" {
			continue
		}
		str.WriteString(cluster + "\n")
		for article, count := range value {
			str.WriteString(fmt.Sprintf("  %v %v\n", article, count))
		}
	}

	stocks, lostWarehouses, err := wb_stocks_analyze.GetStocks(WbKey)
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, fmt.Sprintf("Ошибка при анализе остатков: %v", err))
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
			return
		}
		return
	}

	filePath, err := generateExcel(orders, stocks, K, "wb")
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, fmt.Sprintf("Ошибка при генерации экселя: %v", err))
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
			return
		}
		return
	}

	err = sendMediaMessage(ctx, bot, chatId, filePath)
	if err != nil {
		log.Println("Ошибка отправки сообщения:", err)
		return
	}
	os.Remove(filePath)

	if len(lostWarehouses) > 0 {
		warehousesStr := strings.Builder{}

		for warehouse := range lostWarehouses {
			warehousesStr.WriteString(warehouse + "\n")
		}
		_, err := sendTextMessage(ctx, bot, chatId, fmt.Sprintf("Нужно добавить:\n"+warehousesStr.String()))
		if err != nil {
			return
		}
	}

}

func getYandexFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	text := fmt.Sprintf("Подготовка файла Яндекс")
	message, err := sendTextMessage(ctx, bot, chatId, text)
	if err != nil {
		return
	}

	yandexToken, err := initEnv("variables.env", "yandexToken")
	if err != nil {
		log.Panic(err)
	}

	err = yandex_stickers_fbs.GetOrdersInfo(yandexToken, supplyId)
	if err != nil {
		_, err := sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		filePath := fmt.Sprintf("%v.pdf", yandex_stickers_fbs.DirectoryPath+supplyId)
		sendMediaMessage(ctx, bot, chatId, filePath)
		yandex_stickers_fbs.CleanFiles(supplyId)
	}

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

func sendTextMessage(ctx context.Context, bot *botlib.Bot, chatId int64, text string) (*models.Message, error) {
	message, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text})
	if err != nil {
		return nil, err
	}
	return message, nil
}
func sendMediaMessage(ctx context.Context, bot *botlib.Bot, chatId int64, filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	inputFile := models.InputFileUpload{
		Filename: filePath,
		Data:     file,
	}

	_, err = bot.SendDocument(ctx, &botlib.SendDocumentParams{ChatID: chatId, Document: &inputFile})
	if err != nil {
		return err
	}
	return nil
}

func initEnv(path, name string) (string, error) {
	err := godotenv.Load(path)
	if err != nil {
		log.Printf("Ошибка загрузки файла %s: %v\n", path, err)
		return "", fmt.Errorf("ошибка загрузки файла " + path)
	}
	// Получаем значения переменных среды
	env := os.Getenv(name)

	if env == "" {
		return "", fmt.Errorf("переменная среды " + name + " не установлена")
	}
	return env, err
}
func generateExcel(postings map[string]map[string]int, stocks map[string]map[string]int, K float64, mp string) (string, error) {
	file := excelize.NewFile()
	sheetName := "StocksFBO Analysis"
	file.SetSheetName("Sheet1", sheetName)

	// Заголовки
	headers := []string{"Кластер", "Артикул", "Заказано", "Остатки"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		file.SetCellValue(sheetName, cell, h)
	}

	articles := make(map[string]struct{})

	// Собираем все уникальные артикулы
	for _, postingsMap := range postings {
		for article := range postingsMap {
			articles[article] = struct{}{}
		}
	}
	for _, stocksMap := range stocks {
		for article := range stocksMap {
			articles[article] = struct{}{}
		}
	}

	row := 2
	for cluster, postingsMap := range postings {

		for article := range articles {
			postingCount := postingsMap[article]
			stock := 0
			if clusterStocks, stocksExists := stocks[cluster]; stocksExists {
				stock = clusterStocks[article]
			}

			if stock == 0 || (postingCount > 0 && float64(stock)/float64(postingCount) < K) {
				file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
				file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
				file.SetCellValue(sheetName, "C"+strconv.Itoa(row), postingCount)
				file.SetCellValue(sheetName, "D"+strconv.Itoa(row), stock)
				row++
			}
		}
	}

	// Сохраняем файл
	filePath := mp + "_stock_analysis.xlsx"
	if err := file.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func (m *Manager) AnalyzeStocks(apiKey string, ctx context.Context, b *botlib.Bot) error {
	stocksFBO, err := wb.GetStockFbo(apiKey)
	if err != nil {
		return err
	}

	if stocksFBO == nil {
		return errors.New("newStocks nil")
	}

	type customStock struct {
		stockFBO int
		stockFBS int
	}

	stocksMap := make(map[string]customStock)

	// Заполнение мапы артикулов
	for i := range stocksFBO {
		if stock, hasArticle := stocksMap[stocksFBO[i].SupplierArticle]; hasArticle {
			stock.stockFBO += stocksFBO[i].Quantity
		} else {
			stock.stockFBO = stocksFBO[i].Quantity
		}
	}

	for article, newStocks := range stocksMap {
		var stocksDB []db.Stock
		// Смотрим есть ли артикул в бд
		result := m.db.Where("article = ?", article).Find(&stocksDB)
		if result.Error != nil {
			log.Println("Error finding stocksDB:", result.Error)
		}

		// если артикула нет - заполняем бд
		if len(stocksDB) == 0 {
			stock := db.Stock{Article: article, StocksFBO: &newStocks.stockFBO, Date: time.Now()}
			err = m.db.Create(&stock).Error
			if err != nil {
				return err
			}
			continue
		}

		if newStocks.stockFBO == *stocksDB[0].StocksFBO {
			continue
		}

		// Если стало нулем
		if newStocks.stockFBO == 0 && *stocksDB[0].StocksFBO != 0 {
			// Отправляем уведомление
			_, err = b.SendMessage(ctx, &botlib.SendMessageParams{
				ChatID: m.myChatId,
				Text:   fmt.Sprintf("Нужно добавить наличие fbs для %v", article),
			})
			if err != nil {
				return err
			}
		}

		fmt.Println("Обновляем ", stocksDB[0].Article)

		err = m.db.Model(&db.Stock{}).Where("article = ?", stocksDB[0].Article).Updates(db.Stock{
			StocksFBO: &newStocks.stockFBO,
			Date:      time.Now(),
		}).Error
		if err != nil {
			return err
		}

	}

	return nil
}

//func generateExcel(postings map[string]map[string]int, stocks map[string]map[string]int, K float64, mp string) (string, error) {
//	file := excelize.NewFile()
//	sheetName := "StocksFBO Analysis"
//	file.SetSheetName("Sheet1", sheetName)
//
//	// Заголовки
//	headers := []string{"Кластер", "Артикул", "Заказано", "Остатки"}
//	for i, h := range headers {
//		cell := string(rune('A'+i)) + "1"
//		file.SetCellValue(sheetName, cell, h)
//	}
//
//	articles := make(map[string]interface{})
//
//	for _, postingsMap := range postings {
//		for article := range postingsMap {
//			articles[article] = nil
//		}
//	}
//
//	row := 2
//	for cluster, postingsMap := range postings {
//		if postingsMap == nil {
//			continue
//		}
//
//		if clusterStocks, exists := stocks[cluster]; exists {
//			for article, postingCount := range postingsMap {
//				stock := clusterStocks[article]
//
//				if float64(stock)/float64(postingCount) < K {
//					file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
//					file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
//					file.SetCellValue(sheetName, "C"+strconv.Itoa(row), postingCount)
//					file.SetCellValue(sheetName, "D"+strconv.Itoa(row), stock)
//					row++
//				}
//			}
//		} else {
//			for article := range articles {
//				if countFromPostings, exists := postingsMap[article]; !exists {
//					file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
//					file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
//					file.SetCellValue(sheetName, "C"+strconv.Itoa(row), 0)
//					file.SetCellValue(sheetName, "D"+strconv.Itoa(row), 0)
//				} else {
//					file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
//					file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
//					file.SetCellValue(sheetName, "C"+strconv.Itoa(row), countFromPostings)
//					file.SetCellValue(sheetName, "D"+strconv.Itoa(row), 0)
//				}
//
//			}
//			row++
//		}
//	}
//
//	// Сохраняем файл
//	filePath := mp + "_stock_analysis.xlsx"
//	if err := file.SaveAs(filePath); err != nil {
//		return "", err
//	}
//	return filePath, nil
//}
