package bot

import (
	"WildberriesGo_bot/pkg/OZON"
	"WildberriesGo_bot/pkg/WB"
	"WildberriesGo_bot/pkg/YANDEX"
	"context"
	botlib "github.com/go-telegram/bot"
	"gorm.io/gorm"
	"log"
)

type Service struct {
	token         string
	manager       *Manager
	sqlDB         *gorm.DB
	myChatId      string
	ozonService   OZON.Service
	wbService     WB.Service
	yandexService YANDEX.Service
}

func NewBotService(ozonService OZON.Service, wbService WB.Service, yandexService YANDEX.Service, token string, sqlDB *gorm.DB, myChatId string) Service {
	return Service{
		ozonService:   ozonService,
		wbService:     wbService,
		yandexService: yandexService,
		token:         token,
		sqlDB:         sqlDB,
		myChatId:      myChatId,
	}
}

func (s *Service) GetManager() *Manager {
	return s.manager
}

func (s *Service) Start() {
	s.manager = NewBotManager(s.ozonService, s.wbService, s.yandexService, s.sqlDB, s.myChatId)
	opts := []botlib.Option{botlib.WithDefaultHandler(s.manager.DefaultHandler)}
	newBot, _ := botlib.New(s.token, opts...)
	s.manager.SetBot(newBot)
	go func() {
		log.Printf("Бот запущен\n")
		newBot.Start(context.Background())

	}()
	s.manager.RegisterBotHandlers()
}
