package app

import (
	"WildberriesGo_bot/pkg/bot"
	"WildberriesGo_bot/pkg/db"
	"WildberriesGo_bot/pkg/scheduler"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Application struct {
	envPath string
}

func NewApplication(envPath string) Application {
	return Application{envPath: envPath}
}

func (a Application) Start() {
	wbToken, err := initEnv(a.envPath, "API_KEY_WB")
	if err != nil {
		fmt.Printf("%v", err)
	}
	myChatId, err := initEnv(a.envPath, "myChatId")
	if err != nil {
		fmt.Printf("%v", err)
	}
	token, err := initEnv(a.envPath, "token")
	if err != nil {
		fmt.Printf("%v", err)
	}
	//yandexToken, err := initEnv(a.envPath, "yandexToken")
	//if err != nil {
	//	fmt.Printf("%v", err)
	//}
	dsn, err := initEnv(a.envPath, "DSN")
	if err != nil {
		fmt.Printf("%v", err)
	}

	dataBaseService := db.NewDataBaseService(dsn)
	sqlDB, err := dataBaseService.InitDB()
	if err != nil {
		fmt.Printf("%v", err)
	}

	botService := bot.NewBotService(token, sqlDB, myChatId)
	botService.Start()

	schedulerService := scheduler.NewService(botService.GetManager(), wbToken)
	err = schedulerService.Start()
	if err != nil {
		return
	}

	defer func(sqlDB *gorm.DB) {
		database, err := sqlDB.DB()
		if err != nil {
			log.Panic(err)
		}
		err = database.Close()
		if err == nil {
			log.Println("Соединение закрыто")
		} else {
			log.Panic(err)
		}
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
