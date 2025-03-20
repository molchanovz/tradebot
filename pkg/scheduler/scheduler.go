package scheduler

import (
	"WildberriesGo_bot/pkg/bot"
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

type Service struct {
	botManager *bot.Manager
	wbToken    string
}

func NewService(botManager *bot.Manager, wbToken string) Service {
	return Service{botManager: botManager, wbToken: wbToken}
}

func (s *Service) Start() error {
	scheduler := gocron.NewScheduler(time.Local)
	_, err := scheduler.Every(30).Minute().Do(func() {
		err := s.botManager.AnalyzeStocks(s.wbToken, context.Background(), s.botManager.GetBot())
		if err != nil {
			return
		}
	})
	if err != nil {
		return fmt.Errorf("Ошибка при добавлении задачи: %v\n", err)
	}

	go func() {
		scheduler.StartAsync()
		log.Println("Планировщик запущен")
	}()

	return nil
}
