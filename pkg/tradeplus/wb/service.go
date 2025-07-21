package wb

import (
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = ozon.OrdersDaysAgo
)

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

	if cabinet.SheetLink != nil {
		service.spreadsheetId = *cabinet.SheetLink
	}

	return service
}

type Authorization struct {
	Token string
}

func (s Service) GetOrdersManager() OrdersManager {
	return NewOrdersManager(s.Token, s.spreadsheetId, OrdersDaysAgo)
}

func (s Service) GetStickersFbsManager() StickerManager {
	return NewStickerManager(s.Token)
}
