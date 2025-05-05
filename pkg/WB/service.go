package WB

import (
	"tradebot/pkg/OZON"
	"tradebot/pkg/WB/OrdersAndReturns"
	"tradebot/pkg/WB/stickersFbs"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = OZON.OrdersDaysAgo
	spreadsheetId = "1MpJkuAMmHbUIFFs-J7lAHgsn6qXx6nisbN2yhbEza4c"
)

type Service struct {
	clientId, token           string
	ordersAndReturnsWbManager OrdersAndReturns.WbManager
	stickersWbManager         stickersFbs.WbManager
}

func NewService(token string) *Service {
	return &Service{
		token:                     token,
		ordersAndReturnsWbManager: OrdersAndReturns.NewWbManager(token, spreadsheetId, OrdersDaysAgo),
		stickersWbManager:         stickersFbs.NewWbManager(token),
	}
}

func (s Service) GetOrdersAndReturnsManager() OrdersAndReturns.WbManager {
	return s.ordersAndReturnsWbManager
}

func (s Service) GetStickersFbsManager() stickersFbs.WbManager {
	return s.stickersWbManager
}
