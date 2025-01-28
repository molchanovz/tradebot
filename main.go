package main

import (
	"WildberriesGo_bot/DB"
	"WildberriesGo_bot/WB/wb_orders_returns"
	"WildberriesGo_bot/WB/wb_stickers_fbs"
	"WildberriesGo_bot/WB/wb_stocks_analyze"
	"WildberriesGo_bot/YANDEX/yandex_stickers_fbs"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var myChatId string

var yandexToken = "ACMA:fxJtgnlQjQZsjcTkpP3omw0pyyhEbhFMADBjFPRD:c354e75a"

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

	log.Printf("Инициализация базы данных")
	var sqlDB *sql.DB
	sqlDB, err = DB.InitDB()
	if err != nil {
		log.Panic(err)
	}
	opts := []botlib.Option{botlib.WithDefaultHandler(defaultHandler)}
	b, _ := botlib.New(token, opts...)

	b.RegisterHandler(botlib.HandlerTypeMessageText, "/start", botlib.MatchTypePrefix, startHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "START", botlib.MatchTypePrefix, startHandler)

	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "WB", botlib.MatchTypeExact, wbHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "WB_FBS", botlib.MatchTypeExact, wbFbsHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "YANDEX_FBS", botlib.MatchTypeExact, yandexFbsHandler)
	b.RegisterHandler(botlib.HandlerTypeCallbackQueryData, "WB_ORDERS", botlib.MatchTypePrefix, wbOrdersHandler)

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
		var stocksDB []DB.Stock
		// Смотрим есть ли артикул в бд
		result := DB.Database.Where("article = ?", article).Find(&stocksDB)
		if result.Error != nil {
			log.Println("Error finding stocksApi:", result.Error)
		}

		// если артикула нет - заполняем бд
		if len(stocksDB) == 0 {
			stock := DB.Stock{Article: article, Stock: &quantity, Date: time.Now()}
			err := DB.Database.Create(&stock).Error
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

		err := DB.Database.Model(&DB.Stock{}).Where("article = ?", stocksDB[0].Article).Updates(DB.Stock{
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

	var user DB.User
	// Смотрим есть ли артикул в бд
	result := DB.Database.Where("chatId = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding stocksApi:", result.Error)
	}

	switch user.State {
	case DB.DefaultState:
		{
			sendTextMessage(ctx, bot, chatId, "Не понял тебя. Нажми /start еще раз")
		}
	case DB.WaitingWbState:
		{
			getWbFbs(ctx, bot, chatId, message)
		}
	case DB.WaitingYaState:
		{
			ordersIds, err := yandex_stickers_fbs.GetOrdersIds(yandexToken, message)
			if err != nil {
				return
			}
			str := fmt.Sprint(ordersIds)[1 : len(fmt.Sprint(ordersIds))-1]
			sendTextMessage(ctx, bot, chatId, str)
		}
	}

	err := DB.Database.Model(&DB.User{}).Where("chatId = ?", chatId).Updates(DB.User{
		ChatId: chatId,
		State:  DB.DefaultState,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления DefaultState пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен DefaultState", chatId)

}

func startHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	var chatId int64
	var text string
	startMessage := "Выбери маркетплейс для работы"

	var buttonsRow []models.InlineKeyboardButton
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ВБ", CallbackData: "WB"})
	buttonsRow = append(buttonsRow, models.InlineKeyboardButton{Text: "ЯНДЕКС", CallbackData: "YANDEX_FBS"})
	allButtons := [][]models.InlineKeyboardButton{buttonsRow}
	markup := models.InlineKeyboardMarkup{InlineKeyboard: allButtons}

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

	var user DB.User
	// Смотрим есть ли юзер в бд
	result := DB.Database.Where("chatId = ?", chatId).Find(&user)
	if result.Error != nil {
		log.Println("Error finding stocksApi:", result.Error)
	}

	// если юзера нет - заполняем бд
	if user.ChatId == 0 {
		user := DB.User{ChatId: chatId, State: DB.DefaultState}
		err := DB.Database.Create(&user).Error
		if err != nil {
			log.Println("Ошибка создания пользователя: ", err)
		}
		log.Printf("Пользователь %v создан", chatId)
	} else {
		err := DB.Database.Model(&DB.User{}).Where("chatId = ?", chatId).Updates(DB.User{
			ChatId: chatId,
			State:  DB.DefaultState,
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
func wbFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := DB.Database.Model(&DB.User{}).Where("chatId = ?", chatId).Updates(DB.User{
		ChatId: chatId,
		State:  DB.WaitingWbState,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления WaitingWbState пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен WaitingWbState", chatId)

	text := fmt.Sprintf("Отправь мне номер отгрузки")
	sendTextMessage(ctx, bot, chatId, text)

}
func yandexFbsHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	err := DB.Database.Model(&DB.User{}).Where("chatId = ?", chatId).Updates(DB.User{
		ChatId: chatId,
		State:  DB.WaitingYaState,
	}).Error
	if err != nil {
		log.Println("Ошибка обновления WaitingYaState пользователя: ", err)
	}
	log.Printf("У пользователя %v обновлен WaitingYaState", chatId)

	text := fmt.Sprintf("Отправь мне номер отгрузки")
	sendTextMessage(ctx, bot, chatId, text)

}

func getWbFbs(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {
	wildberriesKey, err := initEnv("variables.env", "API_KEY_WB")
	if err != nil {
		return
	}

	text := fmt.Sprintf("Подготовка файла ВБ")
	sendTextMessage(ctx, bot, chatId, text)

	err = wb_stickers_fbs.GetReadyFile(wildberriesKey, supplyId)
	if err != nil {
		sendTextMessage(ctx, bot, chatId, err.Error())
	} else {
		sendMediaMessage(ctx, bot, chatId, supplyId)
		wb_stickers_fbs.Clean_files(supplyId)
	}
}
func wbOrdersHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.CallbackQuery.From.ID

	wildberriesKey, err := initEnv("variables.env", "API_KEY_WB")
	if err != nil {
		return
	}

	err = wb_orders_returns.WriteToGoogleSheets(wildberriesKey)

	if err != nil {
		sendTextMessage(ctx, bot, chatId, err.Error())
	} else {
		sendTextMessage(ctx, bot, chatId, "Данные Wb FBS за вчерашний день внесены")
	}
}

func sendTextMessage(ctx context.Context, bot *botlib.Bot, chatId int64, text string) {
	_, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}
func sendMediaMessage(ctx context.Context, bot *botlib.Bot, chatId int64, supplyId string) {

	file, err := os.Open("WB/wb_stickers_fbs/" + supplyId + ".pdf")
	if err != nil {
		log.Fatal(err)
	}

	inputFile := models.InputFileUpload{
		Filename: supplyId + ".pdf",
		Data:     file,
	}

	_, err = bot.SendDocument(ctx, &botlib.SendDocumentParams{ChatID: chatId, Document: &inputFile})
	if err != nil {
		log.Printf("%v", err)
		return
	}
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
