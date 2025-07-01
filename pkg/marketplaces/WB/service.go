package WB

import (
	"tradebot/pkg/marketplaces/OZON"
	"tradebot/pkg/marketplaces/WB/OrdersAndReturns"
	"tradebot/pkg/marketplaces/WB/stickersFbs"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = OZON.OrdersDaysAgo
	spreadsheetId = "1Rljs-bxCQCP0DnDfqRGw0SKm1OjRXi2EdJrW5j3M5ts"
)

type Service struct {
	clientId, token   string
	ordersWriter      OrdersAndReturns.WbOrdersManager
	stickersWbManager stickersFbs.WbManager
}

func NewService(token string) *Service {
	return &Service{
		ordersWriter:      OrdersAndReturns.NewWbOrdersManager(token, spreadsheetId, OrdersDaysAgo),
		stickersWbManager: stickersFbs.NewWbManager(token),
	}
}

type Authorization struct {
	Token string
}

func (s Service) GetOrdersManager() OrdersAndReturns.WbOrdersManager {
	return s.ordersWriter
}

func (s Service) GetStickersFbsManager() stickersFbs.WbManager {
	return s.stickersWbManager
}
