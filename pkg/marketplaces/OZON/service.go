package OZON

import (
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces/OZON/orders_and_returns"
	"tradebot/pkg/marketplaces/OZON/stickersFBS"
	"tradebot/pkg/marketplaces/OZON/stocks_analyzer"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 1
	spreadsheetId = "1BQt6vsGKqFKZ06V4PaV2hmnbTS8c2sbSf3-hR7Hr320"
)

type Authorization struct {
	ClientId, Token string
}

type Service struct {
	Authorization
	spreadsheetId string
}

func NewService(cabinet db.Cabinet) Service {
	service := Service{
		Authorization: Authorization{
			ClientId: cabinet.ClientId,
			Token:    cabinet.Key,
		},
		spreadsheetId: spreadsheetId,
	}
	return service
}

func (s Service) GetOrdersAndReturnsManager() orders_and_returns.OzonOrdersManager {
	return orders_and_returns.NewOzonOrdersManager(s.ClientId, s.Token, s.spreadsheetId, OrdersDaysAgo)

}

func (s Service) GetStocksManager() stocks_analyzer.OzonManager {
	return stocks_analyzer.NewManager(s.ClientId, s.Token, StocksDaysAgo)

}

func (s Service) GetStickersFBSManager() stickersFBS.OzonManager {
	return stickersFBS.NewOzonManager(s.ClientId, s.Token)

}
