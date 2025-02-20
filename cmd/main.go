package main

import (
	"WildberriesGo_bot/pkg/OZON/ozon_orders_returns"
	"WildberriesGo_bot/pkg/OZON/ozon_stocks"
	"WildberriesGo_bot/pkg/WB/wb_orders_returns"
	"WildberriesGo_bot/pkg/WB/wb_stickers_fbs"
	"WildberriesGo_bot/pkg/WB/wb_stocks_analyze"
	"WildberriesGo_bot/pkg/YANDEX/yandex_orders_returns"
	"WildberriesGo_bot/pkg/YANDEX/yandex_stickers_fbs"
	"WildberriesGo_bot/pkg/database/postgresql"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
)

var myChatId, yandexToken string

func main() {
	//var apiKey string
	var token string
	//TODO вставить apiKey вместо _
	_, err := initEnv("variables.env", "API_KEY_WB")
	if err != nil {
		log.Panic(err)
	}
	myChatId, err = initEnv("variables.env", "myChatId")
	if err != nil {
		log.Panic(err)
	}
	token, err = initEnv("variables.env", "token")
	if err != nil {
		log.Panic(err)
	}
	yandexToken, err = initEnv("variables.env", "yandexToken")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Инициализация базы данных")
	var sqlDB *sql.DB
	sqlDB, err = postgresql.InitDB()
	if err != nil {
		log.Panic(err.Error())
	}
	opts := []botlib.Option{botlib.WithDefaultHandler(defaultHandler)}
	b, _ := botlib.New(token, opts...)

	b.RegisterHandler(botlib.HandlerTypeMessageText, "/start", botlib.MatchTypePrefix, startHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "START", botlib.MatchTypePrefix, startHandler)

	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbHandler, botlib.MatchTypeExact, wbHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexHandler, botlib.MatchTypeExact, yandexHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonHandler, botlib.MatchTypeExact, ozonHandler)

	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbFbsHandler, botlib.MatchTypeExact, wbFbsHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexFbsHandler, botlib.MatchTypeExact, yandexFbsHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackWbOrdersHandler, botlib.MatchTypePrefix, wbOrdersHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackYandexOrdersHandler, botlib.MatchTypePrefix, yandexOrdersHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonOrdersHandler, botlib.MatchTypePrefix, ozonOrdersHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, CallbackOzonStocksHandler, botlib.MatchTypePrefix, ozonStocksHandler)

	//b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "YANDEX_FBS", botlib.MatchTypePrefix, wbOrdersHandler)

	go func() {
		log.Printf("Бот запущен\n")
		b.Start(context.Background())
	}()

	s := gocron.NewScheduler(time.Local)

	//_, err = s.Every(30).Minute().Do(func() {
	//	err := analyzeStocks(apiKey, context.Background(), b)
	//	if err != nil {
	//		log.Printf("Что-то не так при анализе. %v\n", err)
	//		return
	//	}
	//})
	//if err != nil {
	//	log.Printf("Ошибка при добавлении задачи: %v\n", err)
	//	return
	//}

	go func() {
		log.Printf("Планировщик стартанул!\n")
		s.StartAsync()
	}()

	defer func(sqlDB *sql.DB) {
		err = sqlDB.Close()
		if err != nil {
			log.Panic(err)
		} else {
			log.Println("Соединение закрыто")
		}
	}(sqlDB)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan
	log.Println("Завершение программы...")
}

func analyzeStocks(apiKey string, ctx context.Context, b *botlib.Bot) error {
	var stocksApi []wb_stocks_analyze.Stock
	stocksApi = wb_stocks_analyze.StockFbo(apiKey)

	if stocksApi == nil {
		return errors.New("stocks nil")
	}

	stocksMap := make(map[string]int)

	// Заполнение мапы артикулов
	for i := range stocksApi {
		if _, hasArticle := stocksMap[stocksApi[i].SupplierArticle]; hasArticle {
			stocksMap[stocksApi[i].SupplierArticle] += stocksApi[i].Quantity
		} else {
			stocksMap[stocksApi[i].SupplierArticle] = stocksApi[i].Quantity
		}
	}

	for article, quantity := range stocksMap {
		var stocksDB []postgresql.Stock
		// Смотрим есть ли артикул в бд
		result := postgresql.Database.Where("article = ?", article).Find(&stocksDB)
		if result.Error != nil {
			log.Println("Error finding stocksApi:", result.Error)
		}

		// если артикула нет - заполняем бд
		if len(stocksDB) == 0 {
			stock := postgresql.Stock{Article: article, Stock: &quantity, Date: time.Now()}
			err := postgresql.Database.Create(&stock).Error
			if err != nil {
				return err
			}
			continue
		}

		if *stocksDB[0].Stock == quantity {
			continue
		}
		// Если стало нулем
		if quantity == 0 {
			// Отправляем уведомление
			_, err := b.SendMessage(ctx, &botlib.SendMessageParams{
				ChatID: myChatId,
				Text:   fmt.Sprintf("Нужно добавить наличие fbs для %v", article),
			})
			if err != nil {
				return err
			}
		}

		fmt.Println("Обновляем ", stocksDB[0].Article)

		err := postgresql.Database.Model(&postgresql.Stock{}).Where("article = ?", stocksDB[0].Article).Updates(postgresql.Stock{
			Stock: &quantity,
			Date:  time.Now(),
		}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func defaultHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.Message.From.ID
	message := update.Message.Text

	var user postgresql.User
	// Смотрим есть ли артикул в бд
	result := postgresql.Database.Where("chatId = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding stocksApi:", result.Error)
	}

	switch user.State {
	case postgresql.DefaultState:
		{
			sendTextMessage(ctx, bot, chatId, "Не понял тебя. Нажми /start еще раз")
		}
	case postgresql.WaitingWbState:
		{
			getWbFbs(ctx, bot, chatId, message)
		}
	case postgresql.WaitingYaState:
		{
			getYandexFbs(ctx, bot, chatId, message)
		}
	}

	err := postgresql.Database.Model(&postgresql.User{}).Where("chatId = ?", chatId).Updates(postgresql.User{
		ChatId: chatId,
		State:  postgresql.DefaultState,
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

func startHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
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

	var user postgresql.User
	// Смотрим есть ли юзер в бд
	result := postgresql.Database.Where("chatId = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding stocksApi:", result.Error)
	}

	// если юзера нет - заполняем бд
	if user.ChatId == 0 {
		user := postgresql.User{ChatId: chatId, State: postgresql.DefaultState}
		err := postgresql.Database.Create(&user).Error
		if err != nil {
			log.Println("Ошибка создания пользователя: ", err)
		}
		log.Printf("Пользователь %v создан", chatId)
	} else {
		err := postgresql.Database.Model(&postgresql.User{}).Where("chatId = ?", chatId).Updates(postgresql.User{
			ChatId: chatId,
			State:  postgresql.DefaultState,
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
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Этикетки FBS", CallbackData: "WB_FBS"})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "Вчерашние заказы", CallbackData: "WB_ORDERS"})

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

func wbFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := postgresql.Database.Model(&postgresql.User{}).Where("chatId = ?", chatId).Updates(postgresql.User{
		ChatId: chatId,
		State:  postgresql.WaitingWbState,
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
func yandexFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := postgresql.Database.Model(&postgresql.User{}).Where("chatId = ?", chatId).Updates(postgresql.User{
		ChatId: chatId,
		State:  postgresql.WaitingYaState,
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
		_, err = sendTextMessage(ctx, bot, chatId, "Заказы озон за вчерашний день были внесены")
		if err != nil {
			log.Println(err)
			return
		}
	}
}
func yandexOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := yandex_orders_returns.WriteToGoogleSheets(yandexToken)
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

	filePath, err := generateExcel(postings, stocks, K)
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

func getYandexFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	text := fmt.Sprintf("Подготовка файла Яндекс")
	message, err := sendTextMessage(ctx, bot, chatId, text)
	if err != nil {
		return
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

func generateExcel(postings map[string]map[string]int, stocks map[string]map[string]int, K float64) (string, error) {
	file := excelize.NewFile()
	sheetName := "Stock Analysis"
	file.SetSheetName("Sheet1", sheetName)

	// Заголовки
	headers := []string{"Кластер", "Артикул", "Заказано", "Остатки"}
	for i, h := range headers {
		cell := string(rune('A'+i)) + "1"
		file.SetCellValue(sheetName, cell, h)
	}

	row := 2
	for cluster, postingsMap := range postings {
		if postingsMap == nil {
			continue
		}

		if clusterStocks, exists := stocks[cluster]; exists {
			for article, postingCount := range postingsMap {
				stock := clusterStocks[article]
				if float64(stock)/float64(postingCount) < K {
					file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
					file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
					file.SetCellValue(sheetName, "C"+strconv.Itoa(row), postingCount)
					file.SetCellValue(sheetName, "D"+strconv.Itoa(row), stock)
					row++
				}
			}
		} else {
			for article, postingCount := range postingsMap {
				file.SetCellValue(sheetName, "A"+strconv.Itoa(row), cluster)
				file.SetCellValue(sheetName, "B"+strconv.Itoa(row), article)
				file.SetCellValue(sheetName, "C"+strconv.Itoa(row), postingCount)
				file.SetCellValue(sheetName, "D"+strconv.Itoa(row), 0)
				row++
			}
		}
	}

	// Сохраняем файл
	filePath := "stock_analysis.xlsx"
	if err := file.SaveAs(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}
