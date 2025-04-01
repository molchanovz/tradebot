package WB

import (
	"WildberriesGo_bot/pkg/OZON"
	"WildberriesGo_bot/pkg/WB/OrdersAndReturns"
	"WildberriesGo_bot/pkg/WB/StickersFbs"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = OZON.OrdersDaysAgo
	spreadsheetId = "1jD4ah-kF-3rbskHPXv_nPofoX4Qd73LrzzJ1CQjKKHA"
)

type Service struct {
	clientId, token           string
	ordersAndReturnsWbManager OrdersAndReturns.WbManager
	stickersWbManager         StickersFbs.WbManager
}

func NewService(token string) *Service {
	return &Service{
		token:                     token,
		ordersAndReturnsWbManager: OrdersAndReturns.NewWbManager(token, spreadsheetId, OrdersDaysAgo),
		stickersWbManager:         StickersFbs.NewWbManager(token),
	}
}

func (s Service) GetOrdersAndReturnsManager() OrdersAndReturns.WbManager {
	return s.ordersAndReturnsWbManager
}

func (s Service) GetStickersFbsManager() StickersFbs.WbManager {
	return s.stickersWbManager
}
