package WB

import (
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces"
	"tradebot/pkg/marketplaces/OZON"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = OZON.OrdersDaysAgo
)

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
