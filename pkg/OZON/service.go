package OZON

import (
	"WildberriesGo_bot/pkg/OZON/StickersFBS"
	"WildberriesGo_bot/pkg/OZON/orders_and_returns"
	"WildberriesGo_bot/pkg/OZON/stocks_analyzer"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 2
	spreadsheetId = "15Uq-DMvY61BLd_Y_e1nWKgTPM4LZihwLUDxBv2Sqs5s"
)

type Service struct {
	clientId, token         string
	ordersAndReturnsManager orders_and_returns.OzonManager
	stocksManager           stocks_analyzer.OzonManager
	stickersFbsManager      stickersFBS.OzonManager
}

func NewService(clientId, token string) *Service {
	return &Service{
		clientId:                clientId,
		token:                   token,
		ordersAndReturnsManager: orders_and_returns.NewOzonManager(clientId, token, spreadsheetId, OrdersDaysAgo),
		stocksManager:           stocks_analyzer.NewManager(clientId, token, StocksDaysAgo),
		stickersFbsManager:      stickersFBS.NewOzonManager(clientId, token),
	}
}

func (s Service) GetOrdersAndReturnsManager() orders_and_returns.OzonManager {
	return s.ordersAndReturnsManager

}
func (s Service) GetStocksManager() stocks_analyzer.OzonManager {
	return s.stocksManager

}
func (s Service) GetStickersFBSManager() stickersFBS.OzonManager {
	return s.stickersFbsManager

}
