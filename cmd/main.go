package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tradebot/pkg/bot"
	"tradebot/pkg/db"
)

func main() {
	app := NewApplication()
	app.Start()
}

type Application struct {
	Config struct {
		Database *pg.Options
		Bot      bot.Config
	}
}

func NewApplication() *Application {
	a := new(Application)
	_, err := toml.DecodeFile("cfg/local.toml", &a.Config)
	if err != nil {
		log.Fatal(err)
	}

	return a
}

func (a Application) Start() {
	pgdb := pg.Connect(a.Config.Database)
	dbc := db.New(pgdb)
	err := dbc.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	botService := bot.NewBotService(a.Config.Bot, dbc)
	botService.Start()

	//schedulerService := scheduler.NewService(botService.StickersManager(), wbToken)
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
	}(pgdb)

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
