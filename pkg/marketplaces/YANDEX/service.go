package YANDEX

import (
	"tradebot/pkg/marketplaces/OZON"
)

const (
	spreadsheetId = "1LNNJaqHzLd78BU_N9VBDB-3n6EhFkWA7dzJPU3ysRIw"
	daysAgo       = OZON.OrdersDaysAgo
)

type Service struct {
	ordersAndReturnsManager YandexOrdersManager
	stickersFbsManager      *Manager
}

func NewService(yandexCampaignIdFBO, yandexCampaignIdFBS, token string) *Service {
	return &Service{
		ordersAndReturnsManager: NewYandexOrdersManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token, spreadsheetId, daysAgo),
		stickersFbsManager:      NewManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token),
	}
}

func (s Service) GetOrdersAndReturnsManager() YandexOrdersManager {
	return s.ordersAndReturnsManager
}

func (s Service) GetStickersFbsManager() *Manager {
	return s.stickersFbsManager
}
