package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
	"tradebot/pkg/tradeplus/wb"

	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xuri/excelize/v2"
)

const (
	CallbackWbHandler        = "WB"
	CallbackWbFbsHandler     = "WB-FBS"
	CallbackWbOrdersHandler  = "WB-ORDERS"
	CallbackWbStocksHandler  = "WB-STOCKS"
	CallbackWbReturnsHandler = "WB-RETURNS"
)

func wbHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID
	messageID := update.CallbackQuery.Message.Message.ID

	text := "Кабинет ВБ"

	var buttonsRow []models.InlineKeyboardButton
	var allButtons [][]models.InlineKeyboardButton

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: fmt.Sprintf("%v", CallbackWbFbsHandler)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Анализ заказов", CallbackData: fmt.Sprintf("%v", CallbackWbStocksHandler)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Возвраты в ПВЗ", CallbackData: fmt.Sprintf("%v", CallbackWbReturnsHandler)})
	allButtons = append(allButtons, buttonsRow)
	buttonsRow = []models.InlineKeyboardButton{}

	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})
	allButtons = append(allButtons, buttonsRow)

	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{ChatID: chatID, MessageID: messageID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) stickersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	user, err := m.bl.UserByChatID(ctx, chatID)
	if err != nil {
		log.Println("Ошибка получения пользователя: ", err)
		return
	}

	_, err = m.bl.SetUserStatus(ctx, user, db.StatusWaitingWbState)
	if err != nil {
		log.Println("Ошибка обновления WaitingWbState пользователя: ", err)
		return
	}

	text := "Отправь мне номер отгрузки"

	var buttonBack []models.InlineKeyboardButton

	buttonBack = append(buttonBack, models.InlineKeyboardButton{Text: "Назад", CallbackData: CallbackStartHandler})

	allButtons := [][]models.InlineKeyboardButton{buttonBack}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

	_, err = bot.EditMessageText(ctx, &botlib.EditMessageTextParams{MessageID: update.CallbackQuery.Message.Message.ID, ChatID: chatID, Text: text, ReplyMarkup: markup})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (m *Manager) getWbStickers(ctx context.Context, bot *botlib.Bot, chatID int64, supplyID string) error {
	done := make(chan []string)
	progressChan := make(chan tradeplus.Progress)
	errChan := make(chan error)

	defer tradeplus.CleanFiles()

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		return err
	}

	go func() {
		filePath, err := wb.NewService(cabinets[0]).GetStickersFbsManager().GetReadyFile(supplyID, progressChan)
		if err != nil {
			log.Println("Ошибка при получении файла:", err)
			errChan <- err
			return
		}
		done <- filePath
	}()

	err = WaitReadyFile(ctx, bot, chatID, progressChan, done, errChan)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) wbOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		log.Println(err)
		return
	}

	err = wb.NewService(cabinets[0]).GetOrdersManager().Write()
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
	}

	date := time.Now().AddDate(0, 0, -ozon.OrdersDaysAgo)
	_, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Заказы вб за %v были внесены", date))
	if err != nil {
		log.Println(err)
		return
	}

}
func (m *Manager) wbStocksHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	daysAgo := 14

	chatID := update.CallbackQuery.From.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		m.sl.Errorf("%v", err)
		return
	}

	orders, err := wb.GetOrders(cabinets[0].Key, daysAgo)
	if err != nil {
		m.sl.Errorf("%v", err)
		return
	}

	stocks, lostWarehouses, err := wb.GetStocks(cabinets[0].Key)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Ошибка при анализе остатков: %w", err))
		if err != nil {
			m.sl.Errorf("send msg failed: %v", err)
			return
		}
		return
	}

	filePath, err := generateExcelWB(orders, stocks, db.MarketWB)
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Ошибка при генерации экселя: %w", err))
		if err != nil {
			m.sl.Errorf("send msg failed: %v", err)
			return
		}
		return
	}

	err = SendMediaMessage(ctx, bot, chatID, filePath)
	if err != nil {
		m.sl.Errorf("send media failed: %v", err)
		return
	}
	os.Remove(filePath)

	if len(lostWarehouses) > 0 {
		warehousesStr := strings.Builder{}

		for warehouse := range lostWarehouses {
			warehousesStr.WriteString(warehouse + "\n")
		}
		_, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Нужно добавить: %v\n", warehousesStr.String()))
		if err != nil {
			return
		}
	}
}

func (m *Manager) returnsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatID := update.CallbackQuery.From.ID

	cabinets, err := m.bl.GetCabinetsByMp(ctx, db.MarketWB)
	if err != nil {
		m.sl.Errorf("%v", err)
		return
	}

	filePath, err := wb.NewReturnsManager(cabinets[0].Key).WriteReturns()
	if err != nil {
		_, err = SendTextMessage(ctx, bot, chatID, fmt.Sprintf("Ошибка при анализе остатков: %w", err))
		if err != nil {
			m.sl.Errorf("send msg failed: %v", err)
			return
		}
		return
	}

	defer os.Remove(filePath)

	err = SendMediaMessage(ctx, bot, chatID, filePath)
	if err != nil {
		m.sl.Errorf("send media failed: %v", err)
		return
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
	//			ChatID:    m.myChatID,
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

func generateExcelWB(postings map[string]map[string]int, stocks map[string]map[string]int, mp string) (string, error) {
	file := excelize.NewFile()
	sheetName := "StocksFBO Analysis"
	err := file.SetSheetName("Sheet1", sheetName)
	if err != nil {
		return "", err
	}

	// Заголовки
	headers := []string{"Кластер", "Артикул", "Заказано", "Остатки"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		err = file.SetCellValue(sheetName, cell, h)
		if err != nil {
			return "", err
		}
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

			err = file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
			if err != nil {
				return "", err
			}
			err = file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
			if err != nil {
				return "", err
			}
			err = file.SetCellValue(sheetName, "C"+strconv.Itoa(row), postingCount)
			if err != nil {
				return "", err
			}
			err = file.SetCellValue(sheetName, "D"+strconv.Itoa(row), stock)
			if err != nil {
				return "", err
			}
			row++
		}
	}

	opt := []excelize.AutoFilterOptions{{
		Column:     "",
		Expression: "",
	}}

	rangeRef := fmt.Sprintf("A1:A%v", row)

	err = file.AutoFilter(sheetName, rangeRef, opt)
	if err != nil {
		return "", err
	}

	// Сохраняем файл
	filePath := mp + "_stock_analysis.xlsx"
	if err = file.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}
