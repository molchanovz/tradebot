package OZON

import (
	"tradebot/pkg/marketplaces/OZON/orders_and_returns"
	"tradebot/pkg/marketplaces/OZON/stickersFBS"
	"tradebot/pkg/marketplaces/OZON/stocks_analyzer"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 1
	spreadsheetId = "1BQt6vsGKqFKZ06V4PaV2hmnbTS8c2sbSf3-hR7Hr320"
)

type Service struct {
	ordersWriter       orders_and_returns.OzonOrdersManager
	stocksManager      stocks_analyzer.OzonManager
	stickersFbsManager stickersFBS.OzonManager
}

func NewService(clientId, token string) *Service {
	return &Service{
		ordersWriter:       orders_and_returns.NewOzonOrdersManager(clientId, token, spreadsheetId, OrdersDaysAgo),
		stocksManager:      stocks_analyzer.NewManager(clientId, token, StocksDaysAgo),
		stickersFbsManager: stickersFBS.NewOzonManager(clientId, token),
	}
}

type Authorization struct {
	ClientId, Token string
}

func (s Service) GetOrdersAndReturnsManager() orders_and_returns.OzonOrdersManager {
	return s.ordersWriter

}
func (s Service) GetStocksManager() stocks_analyzer.OzonManager {
	return s.stocksManager

}
func (s Service) GetStickersFBSManager() stickersFBS.OzonManager {
	return s.stickersFbsManager

}
