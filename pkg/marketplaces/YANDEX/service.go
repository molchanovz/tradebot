package YANDEX

import (
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/YANDEX/OrdersAndReturns"
	"tradebot/pkg/marketplaces/YANDEX/yandex_stickers_fbs"
)

const (
	spreadsheetId = "13jiTQsS0QEWqYAvlD54ovtIgYrc1OM9OCGWeJ27rjGw"
	daysAgo       = OZON.OrdersDaysAgo
)

type Service struct {
	ordersAndReturnsManager *OrdersAndReturns.Manager
	stickersFbsManager      *yandex_stickers_fbs.Manager
}

func NewService(yandexCampaignIdFBO, yandexCampaignIdFBS, token string) *Service {
	return &Service{
		ordersAndReturnsManager: OrdersAndReturns.NewManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token, spreadsheetId, daysAgo),
		stickersFbsManager:      yandex_stickers_fbs.NewManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token),
	}
}

func (s Service) GetOrdersAndReturnsManager() *OrdersAndReturns.Manager {
	return s.ordersAndReturnsManager
}

func (s Service) GetStickersFbsManager() *yandex_stickers_fbs.Manager {
	return s.stickersFbsManager
}
