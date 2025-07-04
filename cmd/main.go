package main

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tradebot/pkg/bot"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/YANDEX"
)

func main() {
	app := NewApplication(".env")
	app.Start()
}

type Application struct {
	envPath string
}

func NewApplication(envPath string) Application {
	return Application{envPath: envPath}
}

func (a Application) Start() {
	myChatId, err := initEnv(a.envPath, "myChatId")
	if err != nil {
		fmt.Printf("%v", err)
	}

	dsn, err := initEnv(a.envPath, "DSN")
	if err != nil {
		fmt.Printf("%v", err)
	}
	repo, err := db.NewRepo(dsn)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	yandexCampaignIdFBO, err := initEnv(a.envPath, "yandexCampaignIdFBO")
	if err != nil {
		fmt.Printf("%v", err)
	}
	yandexCampaignIdFBS, err := initEnv(a.envPath, "yandexCampaignIdFBS")
	if err != nil {
		fmt.Printf("%v", err)
	}
	yandexToken, err := initEnv(a.envPath, "yandexToken")
	if err != nil {
		fmt.Printf("%v", err)
	}
	yandexService := YANDEX.NewService(yandexCampaignIdFBO, yandexCampaignIdFBS, yandexToken)

	botToken, err := initEnv(a.envPath, "token")
	if err != nil {
		fmt.Printf("%v", err)
	}
	botService := bot.NewBotService(*yandexService, botToken, repo, myChatId)
	botService.Start()

	//schedulerService := scheduler.NewService(botService.Manager(), wbToken)
	//err = schedulerService.Start()
	//if err != nil {
	//	log.Println(err)
	//	return
	//}

	defer func(db *pg.DB) {
		err := db.Close()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("Соединение закрыто")
	}(repo.DB)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	<-stopChan
	log.Println("Завершение программы...")
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
