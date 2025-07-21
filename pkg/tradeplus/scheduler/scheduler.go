package scheduler

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"time"
	"tradebot/pkg/bot"
)

type Service struct {
	botManager *bot.Manager
	wbToken    string
}

func NewService(botManager *bot.Manager, wbToken string) Service {
	return Service{botManager: botManager, wbToken: wbToken}
}

func (s *Service) Start() error {
	sc := gocron.NewScheduler(time.Local)
	_, err := sc.Every(30).Minute().Do(func() error {
		err := s.botManager.AnalyzeStocks(s.wbToken, context.Background(), s.botManager.GetBot())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Ошибка при добавлении задачи: %v\n", err)
	}

	go func() {
		sc.StartAsync()
		log.Println("Планировщик запущен")
	}()

	return nil
}
