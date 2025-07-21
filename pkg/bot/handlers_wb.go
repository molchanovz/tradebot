package bot

import (
	"context"
	"fmt"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
	"tradebot/pkg/tradeplus/wb"
)

const (
	CallbackWbHandler       = "WB"
	CallbackWbFbsHandler    = "WB-FBS"
	CallbackWbOrdersHandler = "WB-ORDERS"
	CallbackWbStocksHandler = "WB-STOCKS"
)

func wbHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID
	messageId := update.CallbackQuery.Message.Message.ID

	text := "Кабинет ВБ"

	var buttonsRow, buttonBack []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: CallbackWbFbsHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: CallbackWbOrdersHandler})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Остатки", CallbackData: CallbackWbStocksHandler})

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonsRow, buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatId, MessageID: messageId, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}

}

func (m *Manager) stickersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	user, err := m.bl.UserByChatID(ctx, chatId)
	if err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		return
	}

	_, err = m.bl.SetUserStatus(ctx, user, db.StatusWaitingWbState)
	if err != nil {
		log.Println("Ошибка обновления WaitingWbState пользователя: ", err)
		return
	}

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

func (m *Manager) getWbStickers(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	done := make(chan []string)
	progressChan := make(chan tradeplus.Progress)
	errChan := make(chan error)

	defer tradeplus.CleanFiles()

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		filePath, err := wb.NewService(cabinets[0]).GetStickersFbsManager().GetReadyFile(supplyId, progressChan)
		if err != nil {
			log.Println("Ошибка при получении файла:", err)
			errChan <- err
			return
		}
		done <- filePath
	}()

	err = WaitReadyFile(ctx, bot, chatId, progressChan, done, errChan)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(fmt.Sprintf("ошибка отправки сообщения %v", err))
			return
		}
		return
	}

}

func (m *Manager) wbOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		log.Println(err)
		return
	}

	err = wb.NewService(cabinets[0]).GetOrdersManager().Write()
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		date := time.Now().AddDate(0, 0, -ozon.OrdersDaysAgo)
		_, err = SendTextMessage(ctx, bot, chatId, fmt.Sprintf("Заказы вб за %v были внесены", date))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
func (m *Manager) wbStocksHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	daysAgo := 14
	K := 100.0

	chatId := update.CallbackQuery.From.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		log.Println(err)
		return
	}

	orders := wb.GetOrders(cabinets[0].Key, daysAgo)

	stocks, lostWarehouses, err := wb.GetStocks(cabinets[0].Key)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatId, fmt.Sprintf("Ошибка при анализе остатков: %v", err))
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
			return
		}
		return
	}

	filePath, err := generateExcelWB(orders, stocks, K, db.MarketWB)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatId, fmt.Sprintf("Ошибка при генерации экселя: %v", err))
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
			return
		}
		return
	}

	err = SendMediaMessage(ctx, bot, chatId, filePath)
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
		_, err := SendTextMessage(ctx, bot, chatId, fmt.Sprintf("Нужно добавить:\n"+warehousesStr.String()))
		if err != nil {
			return
		}
	}

}

func (m *Manager) AnalyzeStocks(apiKey string, ctx context.Context, b *botlib.Bot) error {
	//stocksFBO, err := wbc.GetStockFbo(apiKey)
	//if err != nil {
	//	return err
	//}
	//
	//if stocksFBO == nil {
	//	return errors.New("newStocks nil")
	//}
	//
	//type customStock struct {
	//	stockFBO int
	//	stockFBS int
	//}
	//
	//stocksMap := make(map[string]customStock)
	//
	//// Заполнение мапы артикулов
	//for i := range stocksFBO {
	//	if stock, hasArticle := stocksMap[stocksFBO[i].SupplierArticle]; hasArticle {
	//		stock.stockFBO += stocksFBO[i].Quantity
	//		stocksMap[stocksFBO[i].SupplierArticle] = stock
	//	} else {
	//		stock := customStock{
	//			stockFBO: stocksFBO[i].Quantity,
	//		}
	//		stocksMap[stocksFBO[i].SupplierArticle] = stock
	//	}
	//}
	//
	//if len(stocksMap) == 0 {
	//	return errors.New("stocksMap nil")
	//}
	//
	//for article, newStocks := range stocksMap {
	//	// Смотрим есть ли артикул в бд
	//	stocks, err := m.repo.GetStocks(article, "wildberries")
	//	if err != nil {
	//		return err
	//	}
	//
	//	// если артикула нет - заполняем бд
	//	if len(stocks) == 0 {
	//		stock := db.Stock{Article: article, CountFbo: &newStocks.stockFBO, UpdatedAt: time.Now(), CabinetID: 0}
	//		err = m.repo.CreateStock(stock)
	//		if err != nil {
	//			return err
	//		}
	//
	//		continue
	//	}
	//
	//	if newStocks.stockFBO == *stocks[0].CountFbo {
	//		continue
	//	}
	//
	//	// Если стало нулем
	//	if newStocks.stockFBO == 0 && *stocks[0].CountFbo != 0 {
	//		// Отправляем уведомление
	//		_, err = b.SendMessage(ctx, &botlib.SendMessageParams{
	//			ChatID:    m.myChatId,
	//			Text:      fmt.Sprintf("На складе <b>WB</b> закончились <code>%v</code>. Проверьте FBS", article),
	//			ParseMode: models.ParseModeHTML,
	//		})
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//	log.Println("Обновляем ", stocks[0].Article)
	//
	//	err = m.repo.UpdateStock(db.Stock{
	//		Article:   stocks[0].Article,
	//		CountFbo:  &newStocks.stockFBO,
	//		UpdatedAt: time.Now(),
	//	})
	//	if err != nil {
	//		return err
	//	}
	//}

	log.Println("НЕ РАБОТАЕТ")

	return nil
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
