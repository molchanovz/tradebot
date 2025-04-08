package OZON

import (
	"WildberriesGo_bot/pkg/OZON/OrdersAndReturns"
	"WildberriesGo_bot/pkg/OZON/StickersFBS"
	"WildberriesGo_bot/pkg/OZON/Stocks"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 2
	spreadsheetId = "15Uq-DMvY61BLd_Y_e1nWKgTPM4LZihwLUDxBv2Sqs5s"
)

type Service struct {
	clientId, token         string
	ordersAndReturnsManager OrdersAndReturns.OzonManager
	stocksManager           Stocks.OzonManager
	stickersFbsManager      StickersFBS.OzonManager
}

func NewService(clientId, token string) *Service {
	return &Service{
		clientId:                clientId,
		token:                   token,
		ordersAndReturnsManager: OrdersAndReturns.NewOzonManager(clientId, token, spreadsheetId, OrdersDaysAgo),
		stocksManager:           Stocks.NewManager(clientId, token, StocksDaysAgo),
		stickersFbsManager:      StickersFBS.NewOzonManager(clientId, token),
	}
}

func (s Service) GetOrdersAndReturnsManager() OrdersAndReturns.OzonManager {
	return s.ordersAndReturnsManager

}

func (s Service) GetStocksManager() Stocks.OzonManager {
	return s.stocksManager

}
func (s Service) GetStickersFBSManager() StickersFBS.OzonManager {
	return s.stickersFbsManager

}
