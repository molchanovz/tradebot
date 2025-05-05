package YANDEX

import (
	"tradebot/pkg/OZON"
	"tradebot/pkg/YANDEX/OrdersAndReturns"
)

const (
	spreadsheetId = "13jiTQsS0QEWqYAvlD54ovtIgYrc1OM9OCGWeJ27rjGw"
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
