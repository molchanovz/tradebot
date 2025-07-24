package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ozonClient "tradebot/pkg/client/ozon"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"

	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xuri/excelize/v2"
)

const (
	CallbackOzonHandler              = "OZON"
	CallbackOzonOrdersHandler        = "OZON-ORDERS_"
	CallbackOzonStocksHandler        = "OZON-STOCKS_"
	CallbackOzonStickersHandler      = "OZON-STICKERS_"
	CallbackOzonPrintStickersHandler = "OZON-PRINT-STICKERS_"
	CallbackOzonCabinetsHandler      = "OZON-CABINETS"
	CallbackSelectOzonCabinetHandler = "CABINET-OZON_"
)

func (m *Manager) ozonHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	messageID := update.CallbackQuery.Message.Message.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketOzon)
	if err != nil {
		log.Println(err)
		return
	}

	text := "Выберите кабинет"
	callbacks := CallbacksForCabinetMarkup{
		PaginationCallback: CallbackOzonCabinetsHandler,
		SelectCallback:     CallbackSelectOzonCabinetHandler,
		BackCallback:       CallbackStartHandler,
	}
	markup := createCabinetsMarkup(cabinets, callbacks, 0, false)

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: messageID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) ozonCabinetHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	messageID := update.CallbackQuery.Message.Message.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetID := parts[1]

	text := "Кабинет Озон"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: fmt.Sprintf("%v%v", CallbackOzonOrdersHandler, cabinetID)})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Анализ", CallbackData: fmt.Sprintf("%v%v", CallbackOzonStocksHandler, cabinetID)})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: fmt.Sprintf("%v%v", CallbackOzonStickersHandler, cabinetID)})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackOzonHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: messageID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) ozonOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketOzon)
	if err != nil {
		log.Println(err)
		return
	}

	titleRange := "!A1"
	fbsRange := "!A2:B1000"
	fboRange := "!D2:E1000"
	returnsRange := "!G2:H1000"

	maxValuesCount, err := ozon.NewService(cabinets[0]).GetOrdersAndReturnsManager().WriteToGoogleSheets(titleRange, fbsRange, fboRange, returnsRange)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, err.Error())
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

	_, err = ozon.NewService(cabinets[1]).GetOrdersAndReturnsManager().WriteToGoogleSheets(titleRange, fbsRange, fboRange, returnsRange)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, err.Error())
		if err != nil {
			log.Printf("%v", err)
			return
		}
		return
	}

	_, err = SendTextMessage(ctx, bot, chatID, "Заказы озон за вчерашний день были внесены")
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) ozonStocksHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetID, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Println("Ошибка конвертации:", err)
		return
	}

	cabinet, err := m.bl.GetCabinetByID(ctx, cabinetID)
	if err != nil {
		log.Println("Ошибка получения кабинета:", err)
		return
	}

	postings := ozon.NewService(cabinet).GetStocksManager().GetPostings()

	stocks := ozon.NewService(cabinet).GetStocksManager().GetStocks()

	filePath, err := generateExcelOzon(postings, stocks, CallbackOzonHandler)
	if err != nil {
		log.Println("Ошибка при создании Excel:", err)
		return
	}

	err = SendMediaMessage(ctx, bot, chatID, filePath)
	if err != nil {
		return
	}

	os.Remove(filePath)
}

// ozonStickersHandler создает клавиатуру кнопок для печати FBS стикеров
func (m *Manager) ozonStickersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	messageID := update.CallbackQuery.Message.Message.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetID := parts[1]

	text := "Печать FBS стикеров. Выберите, какие стикеры распечатать"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Новые", CallbackData: fmt.Sprintf("%v%v_%v", CallbackOzonPrintStickersHandler, cabinetID, ozon.NewLabels)})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Все из сборки", CallbackData: fmt.Sprintf("%v%v_%v", CallbackOzonPrintStickersHandler, cabinetID, ozon.AllLabels)})
	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: fmt.Sprintf("%v%v", CallbackSelectOzonCabinetHandler, cabinetID)})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: messageID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

// ozonPrintStickers точка входа для печати FBS стикеров
func (m *Manager) ozonPrintStickers(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	parts := strings.Split(update.CallbackQuery.Data, "_")
	cabinetID, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Println("Ошибка конвертации:", err)
		return
	}

	flag := parts[2]

	cabinet, err := m.bl.GetCabinetByID(ctx, cabinetID)
	if err != nil {
		log.Println("Ошибка получения кабинета:", err)
		return
	}

	newOrders := ozonClient.PostingslistFbs{}

	printedOrdersMap, err := m.bl.GetPrintedOrders(ctx, cabinet.ID)

	manager := ozon.NewService(cabinet).GetStickersFBSManager(printedOrdersMap)

	var filePaths []string
	done := make(chan []string)
	progressChan := make(chan tradeplus.Progress)
	errChan := make(chan error)

	defer func() {
		close(done)
		close(progressChan)
		close(errChan)
	}()

	switch flag {
	case ozon.AllLabels:
		{
			go func() {
				filePaths, err = manager.GetAllLabels(progressChan)
				if errors.Is(err, ozon.ErrNoRows) {
					_, err = bot.AnswerCallbackQuery(ctx, &botlib.AnswerCallbackQueryParams{Text: "Заказов в сборке нет", ShowAlert: true, CallbackQueryID: update.CallbackQuery.ID})
					if err != nil {
						log.Println(err)
						return
					}
					return
				} else if err != nil {
					log.Println("Ошибка при получении файла:", err)
					errChan <- err
					return
				}

				done <- filePaths
			}()
		}

	case ozon.NewLabels:
		{
			go func() {
				filePaths, newOrders, err = manager.GetNewLabels(progressChan)
				if errors.Is(err, ozon.ErrNoRows) {
					_, err = bot.AnswerCallbackQuery(ctx, &botlib.AnswerCallbackQueryParams{Text: "Новых заказов нет", ShowAlert: true, CallbackQueryID: update.CallbackQuery.ID})
					if err != nil {
						log.Println(err)
						return
					}
					return
				} else if err != nil {
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
			_, err = SendTextMessage(ctx, bot, chatID, err.Error())
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
	}

	err = WaitReadyFile(ctx, bot, chatID, progressChan, done, errChan)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	if flag == ozon.NewLabels && len(newOrders.Result.PostingsFBS) > 0 {
		err = m.bl.CreateOrders(ctx, cabinetID, newOrders)
		if err != nil {
			_, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("ошибка добавления заказов: %v", err))
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
	}

	tradeplus.CleanFiles()
}

func generateExcelOzon(postings map[string]map[string]map[string]int, stocks map[string]map[string]ozon.CustomStocks, mp string) (string, error) {
	file := excelize.NewFile()

	err := createFullStatistic(postings, stocks, file)
	if err != nil {
		return "", err
	}

	for cluster := range postings {
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

func createFullStatistic(postings map[string]map[string]map[string]int, stocks map[string]map[string]ozon.CustomStocks, file *excelize.File) error {
	sheetName := "Общая статистика"
	err := file.SetSheetName("Sheet1", sheetName)
	if err != nil {
		return err
	}

	dates := make([]string, 0, 14)
	for i := 14; i > 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, date)
	}

	headers := []string{"Кластер", "Артикул", "Заказано", "Доступно, шт", "В пути, шт", "Спрос (прогноз)"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		err = file.SetCellValue(sheetName, cell, h)
		if err != nil {
			return err
		}
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

			err = file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
			if err != nil {
				return err
			}

			err = file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
			if err != nil {
				return err
			}

			err = file.SetCellValue(sheetName, "C"+strconv.Itoa(row), totalOrdered)
			if err != nil {
				return err
			}

			err = file.SetCellValue(sheetName, "D"+strconv.Itoa(row), availableStockCount)
			if err != nil {
				return err
			}

			err = file.SetCellValue(sheetName, "E"+strconv.Itoa(row), inWayStockCount)
			if err != nil {
				return err
			}

			err = file.SetCellValue(sheetName, "F"+strconv.Itoa(row), forecast)
			if err != nil {
				return err
			}

			row++
		}
	}

	rangeRef := fmt.Sprintf("A1:F%d", row-1)
	err = file.AutoFilter(sheetName, rangeRef, nil)
	if err != nil {
		return err
	}
	return nil
}

func createStatisticByCluster(cluster string, postings map[string]map[string]map[string]int, stocks map[string]map[string]ozon.CustomStocks, file *excelize.File) error {
	sheetName := cluster
	_, err := file.NewSheet(sheetName)
	if err != nil {
		return err
	}

	headers := []string{"артикул", "имя (необязательно)", "количество"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		err = file.SetCellValue(sheetName, cell, h)
		if err != nil {
			return err
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
			err = file.SetCellValue(sheetName, "A"+strconv.Itoa(row), article)
			if err != nil {
				return err
			}
			err = file.SetCellValue(sheetName, "B"+strconv.Itoa(row), "")
			if err != nil {
				return err
			}
			err = file.SetCellValue(sheetName, "C"+strconv.Itoa(row), forecast-float64(availableStockCount+inWayStockCount))
			if err != nil {
				return err
			}

			row++
		}
	}

	if err = autoFitColumns(file, sheetName, []string{"A", "B", "C"}); err != nil {
		return fmt.Errorf("ошибка автоподбора ширины: %w", err)
	}
	return nil
}
