package OZON

import (
	"errors"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 1
	spreadsheetId = "1WOUHE2qs-c2idJN4pduWkT6PqJzX8XioI-I3ZoeGxMo"
)

var ErrNoRows = errors.New("no rows in result set")

type Service struct {
	marketplaces.Authorization
	spreadsheetId string
}

func NewService(cabinet db.Cabinet) Service {
	service := Service{
		Authorization: marketplaces.Authorization{
			ClientId: *cabinet.ClientID,
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
