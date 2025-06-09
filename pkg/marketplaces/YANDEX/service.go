package YANDEX

import (
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/YANDEX/OrdersAndReturns"
	"tradebot/pkg/marketplaces/YANDEX/yandex_stickers_fbs"
)

const (
	spreadsheetId = "1AdT0oPbO_8enok6ZWJuPXHwMxSb0fHxtaXbGVJ7MqEM"
	daysAgo       = OZON.OrdersDaysAgo
)

type Service struct {
	ordersAndReturnsManager OrdersAndReturns.YandexOrdersManager
	stickersFbsManager      *yandex_stickers_fbs.Manager
}

func NewService(yandexCampaignIdFBO, yandexCampaignIdFBS, token string) *Service {
	return &Service{
		ordersAndReturnsManager: OrdersAndReturns.NewYandexOrdersManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token, spreadsheetId, daysAgo),
		stickersFbsManager:      yandex_stickers_fbs.NewManager(yandexCampaignIdFBO, yandexCampaignIdFBS, token),
	}
}

func (s Service) GetOrdersAndReturnsManager() OrdersAndReturns.YandexOrdersManager {
	return s.ordersAndReturnsManager
}

func (s Service) GetStickersFbsManager() *yandex_stickers_fbs.Manager {
	return s.stickersFbsManager
}
