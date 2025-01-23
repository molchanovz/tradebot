package main

import (
	"WildberriesGo_bot/DB"
	"WildberriesGo_bot/WB/wb_orders_returns"
	"WildberriesGo_bot/WB/wb_stickers_fbs"
	"WildberriesGo_bot/WB/wb_stocks_analyze"
	"context"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	botlib "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

var myChatId string

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

	b, _ := botlib.New(token)
	b.RegisterHandler(botlib.HandlerTypeMessageText, "/start", botlib.MatchTypePrefix, startHandler)
	b.RegisterHandler(botlib.HandlerTypeMessageText, "WB", botlib.MatchTypePrefix, wbHandler)
	b.RegisterHandler(botlib.HandlerTypeMessageText, "/wbOrders", botlib.MatchTypePrefix, wbOrdersHandler)

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
	if err != nil {
		log.Printf("Ошибка при добавлении задачи: %v\n", err)
		return
	}

	go func() {
		log.Printf("Планировщик стартанул!\n")
		s.StartAsync()
	}()

	select {}
}

func analyzeStocks(apiKey string, ctx context.Context, b *botlib.Bot) error {
	var stocksApi []wb_stocks_analyze.Stock
	stocksApi = wb_stocks_analyze.StockFbo(apiKey)

	if stocksApi == nil {
		return errors.New("stocks nil")
	}

	log.Printf("Инициализация базы данных")
	sqlDB, err := DB.InitDB()
	if err != nil {
		return err
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

	err = sqlDB.Close()
	if err != nil {
		return err
	} else {
		log.Println("Соединение закрыто")
	}

	return nil
}

func startHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.Message.From.ID
	name := update.Message.From.FirstName
	text := fmt.Sprintf("Привет, %v", name)
	_, err := bot.SendMessage(ctx, &botlib.SendMessageParams{ChatID: chatId, Text: text})
	if err != nil {
		log.Printf("%v", err)
		return
	}
}
func wbHandler(ctx context.Context, bot *botlib.Bot, update *models.Update) {
	chatId := update.Message.From.ID
	supplyId := update.Message.Text

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
	chatId := update.Message.From.ID

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

	file, err := os.Open("wb_stickers_fbs/" + supplyId + ".pdf")
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
