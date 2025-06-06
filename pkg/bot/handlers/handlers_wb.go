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
	"tradebot/pkg/api/wb"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/WB/stickersFbs"
	"tradebot/pkg/marketplaces/WB/wb_stocks_analyze"
)

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

func (m *Manager) wbFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := m.db.Model(&db.User{}).Where(`"tgId" = ?`, chatId).Updates(db.User{
		TgId:     chatId,
		StatusId: db.WaitingWbState,
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

func (m *Manager) getWbFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	done := make(chan string)
	progressChan := make(chan stickersFbs.Progress)

	var progressMsgId int

	go func() {
		filePath, err := m.wbService.GetStickersFbsManager().GetReadyFile(supplyId, progressChan)
		if err != nil {
			log.Println("Ошибка при получении файла:", err)
			done <- ""
			return
		}
		done <- filePath
	}()

	var lastReportedCurrent int
	var lastTotal int

	for {
		select {
		case progress := <-progressChan:
			if progress.Current != lastReportedCurrent || progress.Total != lastTotal {
				lastReportedCurrent = progress.Current
				lastTotal = progress.Total

				text := fmt.Sprintf("Обработано заказов: %d из %d", progress.Current, progress.Total)

				if progressMsgId == 0 {
					msg, err := bot.SendMessage(ctx, &botlib.SendMessageParams{
						ChatID: chatId,
						Text:   text,
					})
					if err != nil {
						log.Println(err)
					} else {
						progressMsgId = msg.ID
					}
				} else {
					_, err := bot.EditMessageText(ctx, &botlib.EditMessageTextParams{
						ChatID:    chatId,
						MessageID: progressMsgId,
						Text:      text,
					})
					if err != nil {
						log.Println(err)
					}
				}
			}

		case filePath := <-done:

			_, err := bot.SendChatAction(ctx, &botlib.SendChatActionParams{
				ChatID: chatId,
				Action: models.ChatActionUploadDocument,
			})
			if err != nil {
				return
			}

			if filePath == "" {
				_, err = sendTextMessage(ctx, bot, chatId, "Ошибка при получении файла")
				if err != nil {
					log.Println(err)
				}
				return
			}

			filePath = fmt.Sprintf("%v%v.pdf", stickersFbs.WbDirectoryPath, supplyId)
			err = sendMediaMessage(ctx, bot, chatId, filePath)
			if err != nil {
				log.Println(err)
			}

			if progressMsgId != 0 {
				_, err = bot.DeleteMessage(ctx, &botlib.DeleteMessageParams{
					ChatID:    chatId,
					MessageID: progressMsgId,
				})
				if err != nil {
					return
				}
			}

			stickersFbs.CleanFiles(supplyId)
			stickersFbs.CreateDirectories()

			text, markup := createStartAdminMarkup()
			_, err = bot.SendMessage(ctx, &botlib.SendMessageParams{
				ChatID:      chatId,
				Text:        text,
				ReplyMarkup: markup,
			})
			if err != nil {
				log.Printf("%v", err)
			}
			return
		}
	}
}

func (m *Manager) wbOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := m.wbService.GetOrdersManager().WriteToGoogleSheets()
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, err.Error())
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		date := time.Now().AddDate(0, 0, -OZON.OrdersDaysAgo)
		_, err = sendTextMessage(ctx, bot, chatId, fmt.Sprintf("Заказы вб за %v были внесены", date))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
func wbStocksHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	daysAgo := 14
	K := 100.0

	chatId := update.CallbackQuery.From.ID

	WbKey, err := initEnv(".env", "API_KEY_WB")
	if err != nil {
		log.Println(err)
		return
	}

	orders := wb_stocks_analyze.GetOrders(WbKey, daysAgo)

	stocks, lostWarehouses, err := wb_stocks_analyze.GetStocks(WbKey)
	if err != nil {
		_, err = sendTextMessage(ctx, bot, chatId, fmt.Sprintf("Ошибка при анализе остатков: %v", err))
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err)
			return
		}
		return
	}

	filePath, err := generateExcelWB(orders, stocks, K, "wb")
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
			stocksMap[stocksFBO[i].SupplierArticle] = stock
		} else {
			stock := customStock{
				stockFBO: stocksFBO[i].Quantity,
			}
			stocksMap[stocksFBO[i].SupplierArticle] = stock
		}
	}

	if len(stocksMap) == 0 {
		return errors.New("stocksMap nil")
	}

	for article, newStocks := range stocksMap {
		var stocksDB []db.Stock
		// Смотрим есть ли артикул в бд
		result := m.db.Where("article = ? and marketplace = ?", article, "wildberries").Find(&stocksDB)
		if result.Error != nil {
			return result.Error
		}

		// если артикула нет - заполняем бд
		if len(stocksDB) == 0 {
			stock := db.Stock{Article: article, StocksFBO: &newStocks.stockFBO, UpdatedAt: time.Now(), Marketplace: "wildberries"}
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
				ChatID:    m.myChatId,
				Text:      fmt.Sprintf("Нужно добавить наличие <b>WB</b> FBS для <code>%v</code>", article),
				ParseMode: models.ParseModeHTML,
			})
			if err != nil {
				return err
			}
		}

		log.Println("Обновляем ", stocksDB[0].Article)

		err = m.db.Model(&db.Stock{}).Where("article = ? and marketplace = ?", stocksDB[0].Article, "wildberries").Updates(db.Stock{
			StocksFBO: &newStocks.stockFBO,
			UpdatedAt: time.Now(),
		}).Error
		if err != nil {
			return err
		}
	}

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
