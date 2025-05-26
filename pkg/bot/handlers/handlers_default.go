package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"log"
	"math"
	"os"
	"strconv"
	"time"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/OZON/stocks_analyzer"
	"tradebot/pkg/marketplaces/WB"
	"tradebot/pkg/marketplaces/YANDEX"
	"tradebot/pkg/marketplaces/YANDEX/yandex_stickers_fbs"
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
	result := m.db.Where(`"tgId" = ?`, chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	switch user.StatusId {
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
			m.getYandexFbs(ctx, bot, chatId, message)
		}
	default:
		panic("unhandled default case")
	}

	err := m.db.Model(&db.User{}).Where(`"tgId" = ?`, chatId).Updates(db.User{
		TgId:     chatId,
		StatusId: db.EnabledStatus,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления EnabledStatus пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен EnabledStatus", chatId)

}

func createStartAdminMarkup() (string, models.InlineKeyboardMarkup) {
	startMessage := "Выбери маркетплейс для работы"
	var buttonsRow []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", CallbackData: CallbackWbHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", CallbackData: CallbackYandexHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ОЗОН", CallbackData: CallbackOzonHandler})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func createStartUserMarkup() (string, models.InlineKeyboardMarkup) {
	startMessage := "Для доступа к функционалу бота пиши @molchanovz. А пока можешь перейти в наши магазины"
	var buttonsRow []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", URL: "https://www.wildberries.ru/seller/27566"})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", URL: "https://market.yandex.ru/business--metr-v-kube/3697903"})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ОЗОН", URL: "https://www.ozon.ru/seller/metr-v-kube-259267"})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}
	return startMessage, markup
}

func (m *Manager) startHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	var chatId int64
	var text string

	if update.Message != nil {
		chatId = update.Message.From.ID
	} else {
		chatId = update.CallbackQuery.From.ID
	}

	var user db.User
	// Смотрим есть ли юзер в бд
	result := m.db.Where(`"tgId" = ?`, chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding chatId: ", result.Error)
	}

	// если юзера нет - заполняем бд
	if user.TgId == 0 {
		user = db.User{TgId: chatId, StatusId: db.EnabledStatus}
		err := m.db.Create(&user).Error
		if err != nil {
			log.Println("Ошибка создания пользователя: ", err)
		}
		log.Printf("Пользователь %v создан", chatId)
	} else {
		err := m.db.Model(&db.User{}).Where(`"tgId" = ?`, chatId).Updates(db.User{
			StatusId: db.EnabledStatus,
		}).Error
		if err != nil {
			log.Println("Ошибка обновления EnabledStatus пользователя: ", err)
		}
		log.Printf("У пользователя %v обновлен EnabledStatus", chatId)
	}

	var startMessage string
	var markup models.InlineKeyboardMarkup

	if user.IsAdmin {
		startMessage, markup = createStartAdminMarkup()
	} else {
		startMessage, markup = createStartUserMarkup()
	}

	if update.Message != nil {
		name := update.Message.From.FirstName
		text = fmt.Sprintf("Привет, %v. %v", name, startMessage)
		_, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text, ReplyMarkup: markup})
		if err != nil {
			log.Printf("%v", err)
			return
		}

	} else {
		messageId := update.CallbackQuery.Message.Message.ID
		text = startMessage
		_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{MessageID: messageId, ChatID: chatId, Text: text, ReplyMarkup: markup})
		if err != nil {
			log.Printf("%v", err)
			return
		}
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

func (m *Manager) ozonClustersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	clusters := m.ozonService.GetStocksManager().GetClusters()

	fmt.Println(clusters.Clusters)
}

func (m *Manager) getYandexFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	text := fmt.Sprintf("Подготовка файла Яндекс")
	message, err := sendTextMessage(ctx, bot, chatId, text)
	if err != nil {
		return
	}

	err = m.yandexService.GetStickersFbsManager().GetOrdersInfo(supplyId)
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

	opt := []excelize.AutoFilterOptions{{
		Column:     "",
		Expression: "",
	}}

	rangeRef := fmt.Sprintf("A1:A%v", row)

	err := file.AutoFilter(sheetName, rangeRef, opt)
	if err != nil {
		return "", err
	}

	// Сохраняем файл
	filePath := mp + "_stock_analysis.xlsx"
	if err := file.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func generateExcelOzon(postings map[string]map[string]map[string]int, stocks map[string]map[string]stocks_analyzer.CustomStocks, K float64, mp string) (string, error) {
	file := excelize.NewFile()

	err := createFullStatistic(postings, stocks, file)
	if err != nil {
		return "", err
	}

	for cluster, _ := range postings {
		err = createStatisticByCluster(cluster, postings, stocks, file)
		if err != nil {
			return "", err
		}
	}

	filePath := mp + "_stock_analysis.xlsx"
	if err = file.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func createFullStatistic(postings map[string]map[string]map[string]int, stocks map[string]map[string]stocks_analyzer.CustomStocks, file *excelize.File) error {
	sheetName := "Общая статистика"
	file.SetSheetName("Sheet1", sheetName)

	dates := make([]string, 0, 14)
	for i := 14; i > 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, date)
	}

	headers := []string{"Кластер", "Артикул", "Заказано", "Доступно, шт", "В пути, шт", "Спрос (прогноз)"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		file.SetCellValue(sheetName, cell, h)
		err := file.SetColWidth(sheetName, string(rune('A'+i)), string(rune('A'+i)), float64(len(h)))
		if err != nil {
			return fmt.Errorf("ошибка настройки ширины колонки %s: %v", string(rune('A'+i)), err)
		}
	}

	// все уникальные артикулы
	articles := make(map[string]struct{})
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
			salesData := make([]float64, 0, 14)
			totalOrdered := 0

			for _, date := range dates {
				if qty, exists := postingsMap[article][date]; exists {
					salesData = append(salesData, float64(qty))
					totalOrdered += qty
				} else {
					salesData = append(salesData, 0)
				}
			}

			forecast := calculateSmartDemandForecast(salesData)

			availableStockCount := 0
			inWayStockCount := 0
			if clusterStocks, stocksExists := stocks[cluster]; stocksExists {
				if stock, articleExists := clusterStocks[article]; articleExists {
					availableStockCount = stock.AvailableStockCount
					inWayStockCount = stock.TransitStockCount + stock.RequestedStockCount
				}
			}

			file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
			file.SetCellValue(sheetName, "C"+strconv.Itoa(row), totalOrdered)
			file.SetCellValue(sheetName, "D"+strconv.Itoa(row), availableStockCount)
			file.SetCellValue(sheetName, "E"+strconv.Itoa(row), inWayStockCount)
			file.SetCellValue(sheetName, "F"+strconv.Itoa(row), forecast)

			row++
		}
	}

	rangeRef := fmt.Sprintf("A1:F%d", row-1)
	err := file.AutoFilter(sheetName, rangeRef, nil)
	if err != nil {
		return err
	}
	return nil
}
func createStatisticByCluster(cluster string, postings map[string]map[string]map[string]int, stocks map[string]map[string]stocks_analyzer.CustomStocks, file *excelize.File) error {
	sheetName := cluster
	file.NewSheet(sheetName)

	headers := []string{"артикул", "имя (необязательно)", "количество"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		file.SetCellValue(sheetName, cell, h)
	}

	// все уникальные артикулы
	articles := make(map[string]struct{})
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

	dates := make([]string, 0, 14)
	for i := 14; i > 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, date)
	}

	row := 2
	postingsMap := postings[cluster]
	for article := range articles {
		salesData := make([]float64, 0, 14)
		totalOrdered := 0

		for _, date := range dates {
			if qty, exists := postingsMap[article][date]; exists {
				salesData = append(salesData, float64(qty))
				totalOrdered += qty
			} else {
				salesData = append(salesData, 0)
			}
		}

		availableStockCount := 0
		inWayStockCount := 0
		if clusterStocks, stocksExists := stocks[cluster]; stocksExists {
			if stock, articleExists := clusterStocks[article]; articleExists {
				availableStockCount = stock.AvailableStockCount
				inWayStockCount = stock.TransitStockCount + stock.RequestedStockCount
			}
		}

		forecast := calculateSmartDemandForecast(salesData)

		if forecast > float64(availableStockCount+inWayStockCount) && forecast != 0 {
			file.SetCellValue(sheetName, "A"+strconv.Itoa(row), article)
			file.SetCellValue(sheetName, "B"+strconv.Itoa(row), "")
			file.SetCellValue(sheetName, "C"+strconv.Itoa(row), forecast-float64(availableStockCount+inWayStockCount))

			row++
		}

	}

	if err := autoFitColumns(file, sheetName, []string{"A", "B", "C"}); err != nil {
		return fmt.Errorf("ошибка автоподбора ширины: %v", err)
	}
	return nil
}

// Функция для автоподбора ширины колонок
func autoFitColumns(f *excelize.File, sheet string, columns []string) error {
	for _, col := range columns {
		maxWidth := 8.0 // Минимальная ширина по умолчанию
		rows, err := f.GetRows(sheet)
		if err != nil {
			return err
		}

		// Находим максимальную длину содержимого в колонке
		for _, row := range rows {
			colIdx := int(col[0] - 'A')
			if colIdx < len(row) {
				cellValue := row[colIdx]
				// Учитываем длину текста + 2 символа для отступов
				width := float64(len(cellValue))*1.1 + 2
				if width > maxWidth {
					maxWidth = width
				}
			}
		}

		// Устанавливаем ширину
		if err := f.SetColWidth(sheet, col, col, maxWidth); err != nil {
			return err
		}
	}
	return nil
}

func calculateSmartDemandForecast(salesData []float64) float64 {
	if len(salesData) == 0 {
		return 0
	}

	// Настройки
	const (
		shortWindow = 4  // Анализ последних 4 дней для "горячего" тренда
		longWindow  = 14 // Анализ за 14 дней для базового уровня
	)

	// 1. Вычисляем "горячий" тренд (последние 4 дня)
	hotTrend := 0.0
	if len(salesData) >= shortWindow {
		recent := salesData[len(salesData)-shortWindow:]
		first, last := recent[0], recent[len(recent)-1]
		if first > 0 {
			hotTrend = last / first // Рост в последние дни
		}
	}

	// 2. Среднее за весь период (14 дней)
	fullPeriodAverage := mean(salesData)

	// 3. Среднее за последние 4 дня
	recentAverage := 0.0
	if len(salesData) >= shortWindow {
		recentAverage = mean(salesData[len(salesData)-shortWindow:])
	} else {
		recentAverage = mean(salesData)
	}

	// 4. Динамический вес для тренда
	trendWeight := 0.5  // Базовый вес тренда
	if hotTrend > 2.0 { // Если рост более 2x
		trendWeight = 0.8 // Сильнее учитываем тренд
	}

	// 5. Комбинированный прогноз
	forecast := (recentAverage*trendWeight + fullPeriodAverage*(1-trendWeight)) * float64(longWindow)

	// Гарантируем, что прогноз не ниже последних продаж
	if len(salesData) > 0 {
		lastDaySales := salesData[len(salesData)-1]
		minForecast := lastDaySales * float64(longWindow) * 0.7 // Не менее 70% от последнего дня
		if forecast < minForecast {
			forecast = minForecast
		}
	}

	return math.Round(forecast)
}

func mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
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
