package bot

import (
	"context"
	botlib "github.com/go-telegram/bot"
	"log"
	bot "tradebot/pkg/bot/handlers"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/WB"
	"tradebot/pkg/marketplaces/YANDEX"
)

type Service struct {
	token         string
	manager       *bot.Manager
	dbc           *db.Repo
	myChatId      string
	wbService     WB.Service
	yandexService YANDEX.Service
}

func NewBotService(yandexService YANDEX.Service, token string, dbc *db.Repo, myChatId string) Service {
	return Service{
		yandexService: yandexService,
		token:         token,
		dbc:           dbc,
		myChatId:      myChatId,
	}
}

func (s *Service) Manager() *bot.Manager {
	return s.manager
}

func (s *Service) Start() {
	s.manager = bot.NewBotManager(s.wbService, s.yandexService, s.dbc, s.myChatId)
	opts := []botlib.Option{botlib.WithDefaultHandler(s.manager.DefaultHandler)}
	newBot, _ := botlib.New(s.token, opts...)
	s.manager.SetBot(newBot)
	go func() {
		log.Printf("Бот запущен\n")
		newBot.Start(context.Background())

	}()
	s.manager.RegisterBotHandlers()
}
