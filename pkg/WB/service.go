package WB

import (
	"WildberriesGo_bot/pkg/OZON"
	"WildberriesGo_bot/pkg/WB/OrdersAndReturns"
	"WildberriesGo_bot/pkg/WB/StickersFbs"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = OZON.OrdersDaysAgo
	spreadsheetId = "1e5wiZUXTv419NJW_RUFGMNOY2C8KCXTwAXFwQztBhRo"
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
