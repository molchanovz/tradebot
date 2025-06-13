package bot

import (
	"context"
	"errors"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"tradebot/pkg/api/ozon"
	"tradebot/pkg/db"
	"tradebot/pkg/fbsPrinter"
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/OZON/stickersFBS"
	"tradebot/pkg/marketplaces/OZON/stocks_analyzer"
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
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
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
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Printf("%v", err)
			return
		}
		return
	}

	_, err = SendTextMessage(ctx, bot, chatId, "Заказы озон за вчерашний день были внесены")
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

	err = SendMediaMessage(ctx, bot, chatId, filePath)
	if err != nil {
		return
	}

	os.Remove(filePath)

}

// Хендрер для "FBS"
func (m *Manager) ozonStickersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	messageId := update.CallbackQuery.Message.Message.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetId := parts[1]

	var cabinet db.Cabinet

	result := m.db.Where(`"cabinetsId" = ?`, cabinetId).Find(&cabinet)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	text := "Печать FBS стикеров. Выберите, какие стикеры распечатать"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Новые", CallbackData: fmt.Sprintf("%v%v_%v", CallbackOzonPrintStickersHandler, cabinetId, stickersFBS.NewLabels)})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Все из сборки", CallbackData: fmt.Sprintf("%v%v_%v", CallbackOzonPrintStickersHandler, cabinetId, stickersFBS.AllLabels)})
	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: fmt.Sprintf("%v%v", CallbackSelectCabinetHandler, cabinetId)})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

// Хендлер для печати стикеров "FBS"
func (m *Manager) ozonPrintStickers(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetId := parts[1]
	flag := parts[2]
	var err error
	var cabinet db.Cabinet

	result := m.db.Where(`"cabinetsId" = ?`, cabinetId).Find(&cabinet)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	newOrders := ozon.PostingslistFbs{}

	printedOrdersMap := make(map[string]struct{})
	var printedOrders []db.Order

	result = m.db.Where(`"marketplace" = ?`, "ozon").Find(&printedOrders)
	if result.Error != nil {
		log.Println("Error finding user:", result.Error)
	}

	for _, order := range printedOrders {
		printedOrdersMap[order.PostingNumber] = struct{}{}
	}

	manager := OZON.NewService(cabinet).GetStickersFBSManager(printedOrdersMap)

	var filePaths []string
	done := make(chan []string)
	progressChan := make(chan fbsPrinter.Progress)
	errChan := make(chan error)

	switch flag {
	case stickersFBS.AllLabels:
		{
			go func() {
				filePaths, err = manager.GetAllLabels(progressChan)
				if err != nil {
					log.Println("Ошибка при получении файла:", err)
					errChan <- err
					return
				}

				done <- filePaths
			}()

		}

	case stickersFBS.NewLabels:
		{
			go func() {
				filePaths, newOrders, err = manager.GetNewLabels(progressChan)
				if err != nil {
					log.Println("Ошибка при получении файла:", err)
					errChan <- err
					return
				}

				done <- filePaths
			}()
		}

	default:
		err = errors.New("неопознанный флаг для печати")
		if err != nil {
			_, err = SendTextMessage(ctx, bot, chatId, err.Error())
			if err != nil {
				return
			}
			return
		}
	}

	err = WaitReadyFile(ctx, bot, chatId, progressChan, done, errChan)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			return
		}
		return
	}

	if flag == stickersFBS.NewLabels && len(newOrders.Result.PostingsFBS) > 0 {
		orders := make([]db.Order, 0, len(newOrders.Result.PostingsFBS))

		for _, order := range newOrders.Result.PostingsFBS {
			orders = append(orders, db.Order{
				PostingNumber: order.PostingNumber,
				Marketplace:   "ozon",
			})
		}

		result = m.db.Create(orders)
		if result.Error != nil {
			log.Println("Error creating orders:", result.Error)
		}
	}

	fbsPrinter.CleanFiles()
}

func (m *Manager) ozonClustersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	clusters := m.ozonService.GetStocksManager().GetClusters()

	fmt.Println(clusters.Clusters)
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
