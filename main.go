package main

import (
	"WildberriesGo_bot/API"
	"WildberriesGo_bot/DB"
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
	var apiKey string
	var token string
	apiKey, err := initEnv("variables.env", "api_key")
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

	go func() {
		log.Printf("Бот запущен\n")
		b.Start(context.Background())
	}()

	s := gocron.NewScheduler(time.Local)

	_, err = s.Every(30).Minute().Do(func() {
		err := analyzeStocks(apiKey, context.Background(), b)
		if err != nil {
			log.Printf("Что-то не так при анализе. %v\n", err)
			return
		}
	})
	if err != nil {
		log.Printf("Ошибка при добавлении задачи: %v\n", err)
		return
	}

	go func() {
		log.Printf("Планировщик стартанул\n")
		s.StartAsync()
	}()

	select {}
}

func analyzeStocks(apiKey string, ctx context.Context, b *botlib.Bot) error {
	var stocksApi []API.Stock
	stocksApi = API.StockFbo(apiKey)

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

		err = sqlDB.Close()
		if err != nil {
			return err
		}
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
