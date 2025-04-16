package YANDEX

import (
	"WildberriesGo_bot/pkg/OZON"
	"WildberriesGo_bot/pkg/YANDEX/OrdersAndReturns"
)

const (
	spreadsheetId = "15jEh_t2eY3nNQ1fWWieUTnCBUAiULDqDJDDgSjWKu20"
	daysAgo       = OZON.OrdersDaysAgo
)

type Service struct {
	token                   string
	ordersAndReturnsManager *OrdersAndReturns.Manager
}

func NewService(yandexCampaignIdFBO, yandexCampaignIdFBS, token string) *Service {
	return &Service{token: token,
		ordersAndReturnsManager: OrdersAndReturns.NewManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token, spreadsheetId, daysAgo),
	}
}

func (s Service) GetOrdersAndReturnsManager() *OrdersAndReturns.Manager {
	return s.ordersAndReturnsManager

}
