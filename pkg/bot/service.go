package bot

import (
	"context"
	botlib "github.com/go-telegram/bot"
	"gorm.io/gorm"
	"log"
	bot "tradebot/pkg/bot/handlers"
	"tradebot/pkg/marketplaces/WB"
	"tradebot/pkg/marketplaces/YANDEX"
)

type Service struct {
	token         string
	manager       *bot.Manager
	sqlDB         *gorm.DB
	myChatId      string
	wbService     WB.Service
	yandexService YANDEX.Service
}

func NewBotService(wbService WB.Service, yandexService YANDEX.Service, token string, sqlDB *gorm.DB, myChatId string) Service {
	return Service{
		wbService:     wbService,
		yandexService: yandexService,
		token:         token,
		sqlDB:         sqlDB,
		myChatId:      myChatId,
	}
}

func (s *Service) Manager() *bot.Manager {
	return s.manager
}

func (s *Service) Start() {
	s.manager = bot.NewBotManager(s.wbService, s.yandexService, s.sqlDB, s.myChatId)
	opts := []botlib.Option{botlib.WithDefaultHandler(s.manager.DefaultHandler)}
	newBot, _ := botlib.New(s.token, opts...)
	s.manager.SetBot(newBot)
	go func() {
		log.Printf("Бот запущен\n")
		newBot.Start(context.Background())

	}()
	s.manager.RegisterBotHandlers()
}
