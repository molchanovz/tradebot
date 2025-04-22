package bot

import (
	"WildberriesGo_bot/pkg/OZON"
	"WildberriesGo_bot/pkg/OZON/stocks_analyzer"
	"WildberriesGo_bot/pkg/WB"
	"WildberriesGo_bot/pkg/YANDEX"
	"WildberriesGo_bot/pkg/YANDEX/yandex_stickers_fbs"
	"WildberriesGo_bot/pkg/db"
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
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
	CallbackOzonStickersHandler = "OZON_STICKERS"
	CallbackClustersHandler     = "OZON_CLUSTERS"
)

type Manager struct {
	b             *botlib.Bot
	db            *gorm.DB
	ozonService   OZON.Service
	wbService     WB.Service
	yandexService YANDEX.Service
	myChatId      string
}

func NewBotManager(ozonService OZON.Service, wbService WB.Service, yandexService YANDEX.Service, db *gorm.DB, myChatId string) *Manager {
	return &Manager{
		ozonService:   ozonService,
		wbService:     wbService,
		yandexService: yandexService,
		db:            db,
		myChatId:      myChatId,
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
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbOrdersHandler, botlib.MatchTypePrefix, m.wbOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexOrdersHandler, botlib.MatchTypePrefix, m.yandexOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonOrdersHandler, botlib.MatchTypePrefix, m.ozonOrdersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStocksHandler, botlib.MatchTypePrefix, m.ozonStocksHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbStocksHandler, botlib.MatchTypePrefix, wbStocksHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStickersHandler, botlib.MatchTypePrefix, m.ozonStickersHandler)
	m.b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackClustersHandler, botlib.MatchTypePrefix, m.ozonClustersHandler)

	//b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "YANDEX_FBS", botlib.MatchTypePrefix, wbOrdersHandler)

}

func (m *Manager) DefaultHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.Message.From.ID
	message := update.Message.Text

	var user db.User
	// Смотрим есть ли артикул в бд
	result := m.db.Where("tg_id = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	switch user.Status {
	case db.EnabledStatus:
		{
			sendTextMessage(ctx, bot, chatId, "Не понял тебя. Нажми /start еще раз")
		}
	case db.WaitingWbState:
		{
			m.getWbFbs(ctx, bot, chatId, message)
		}
	case db.WaitingYaState:
		{
			getYandexFbs(ctx, bot, chatId, message)
		}
	default:
		panic("unhandled default case")
	}

	err := m.db.Model(&db.User{}).Where("tg_id = ?", chatId).Updates(db.User{
		TgId:   chatId,
		Status: db.EnabledStatus,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления EnabledStatus пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен EnabledStatus", chatId)

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
	result := m.db.Where("tg_id = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding chatId: ", result.Error)
	}

	// если юзера нет - заполняем бд
	if user.TgId == 0 {
		user := db.User{TgId: chatId, Status: db.EnabledStatus}
		err := m.db.Create(&user).Error
		if err != nil {
			log.Println("Ошибка создания пользователя: ", err)
		}
		log.Printf("Пользователь %v создан", chatId)
	} else {
		err := m.db.Model(&db.User{}).Where("tg_id = ?", chatId).Updates(db.User{
			TgId:   chatId,
			Status: db.EnabledStatus,
		}).Error
		if err != nil {
			log.Println("Ошибка обновления EnabledStatus пользователя: ", err)
		}
		log.Printf("У пользователя %v обновлен EnabledStatus", chatId)
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

func (m *Manager) yandexFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := m.db.Model(&db.User{}).Where("tg_id = ?", chatId).Updates(db.User{
		TgId:   chatId,
		Status: db.WaitingYaState,
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

func (m *Manager) ozonClustersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	clusters := m.ozonService.GetStocksManager().GetClusters()

	fmt.Println(clusters.Clusters)
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

func generateExcelWB(postings map[string]map[string]int, stocks map[string]map[string]int, K float64, mp string) (string, error) {
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

			file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
			file.SetCellValue(sheetName, "C"+strconv.Itoa(row), postingCount)
			file.SetCellValue(sheetName, "D"+strconv.Itoa(row), stock)
			row++

		}
	}

	// Сохраняем файл
	filePath := mp + "_stock_analysis.xlsx"
	if err := file.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func generateExcelOzon(postings map[string]map[string]int, stocks map[string]map[string]stocks_analyzer.CustomStocks, K float64, mp string) (string, error) {
	file := excelize.NewFile()
	sheetName := "StocksFBO Analysis"
	file.SetSheetName("Sheet1", sheetName)

	// Заголовки
	headers := []string{"Кластер", "Артикул", "Заказано", "Доступно, шт", "В пути, шт"}
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
			availableStockCount := 0
			inWayStockCount := 0
			if clusterStocks, stocksExists := stocks[cluster]; stocksExists {
				availableStockCount = clusterStocks[article].AvailableStockCount
				inWayStockCount = clusterStocks[article].TransitStockCount + clusterStocks[article].RequestedStockCount
			}

			file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
			file.SetCellValue(sheetName, "C"+strconv.Itoa(row), postingCount)
			file.SetCellValue(sheetName, "D"+strconv.Itoa(row), availableStockCount)
			file.SetCellValue(sheetName, "E"+strconv.Itoa(row), inWayStockCount)
			row++

		}
	}

	// Сохраняем файл
	filePath := mp + "_stock_analysis.xlsx"
	if err := file.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
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
