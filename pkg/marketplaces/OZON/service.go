package OZON

import (
	"tradebot/db"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 1
	spreadsheetId = "1WOUHE2qs-c2idJN4pduWkT6PqJzX8XioI-I3ZoeGxMo"
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

func (s Service) GetOrdersAndReturnsManager() OrdersManager {
	return NewOrdersManager(s.ClientId, s.Token, s.spreadsheetId, OrdersDaysAgo)

}

func (s Service) GetStocksManager() AnalyzeManager {
	return NewAnalyzeManager(s.ClientId, s.Token, StocksDaysAgo)

}

func (s Service) GetStickersFBSManager(printedOrders map[string]struct{}) StickerManager {
	return NewStickerManager(s.ClientId, s.Token, printedOrders)

}
