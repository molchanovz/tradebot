package app

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tradebot/pkg/bot"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/WB"
	"tradebot/pkg/marketplaces/YANDEX"
	"tradebot/pkg/scheduler"
)

type Application struct {
	envPath string
}

func New(envPath string) Application {
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
	dataBaseService := db.NewService(dsn)
	sqlDB, err := dataBaseService.InitDB()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	if sqlDB == nil {
		return
	}

	wbToken, err := initEnv(a.envPath, "API_KEY_WB")
	if err != nil {
		fmt.Printf("%v", err)
	}
	wbService := WB.NewService(wbToken)

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
	botService := bot.NewBotService(*wbService, *yandexService, botToken, sqlDB, myChatId)
	botService.Start()

	schedulerService := scheduler.NewService(botService.Manager(), wbToken)
	err = schedulerService.Start()
	if err != nil {
		log.Println(err)
		return
	}

	defer func(sqlDB *gorm.DB) {

		database, err := sqlDB.DB()
		if err != nil {
			log.Println(err)
			return
		}

		err = database.Close()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("Соединение закрыто")
	}(sqlDB)

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
