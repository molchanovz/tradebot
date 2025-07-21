package ozon

import (
	"errors"
	"tradebot/pkg/tradeplus"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 1
)

var ErrNoRows = errors.New("no rows in result set")

type Service struct {
	tradeplus.Authorization
	spreadsheetId string
}

func NewService(cabinet tradeplus.Cabinet) Service {
	service := Service{
		Authorization: tradeplus.Authorization{
			Token: cabinet.Key,
		},
	}

	if cabinet.ClientID != nil {
		service.ClientId = *cabinet.ClientID
	}

	if cabinet.SheetLink != nil {
		service.spreadsheetId = *cabinet.SheetLink
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
