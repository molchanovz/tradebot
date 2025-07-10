package OZON

import (
	"errors"
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = 1
)

var ErrNoRows = errors.New("no rows in result set")

type Service struct {
	marketplaces.Authorization
	spreadsheetId string
}

func NewService(cabinet db.Cabinet) Service {
	service := Service{
		Authorization: marketplaces.Authorization{
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
