package WB

import (
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces"
	"tradebot/pkg/marketplaces/OZON"
)

const (
	StocksDaysAgo = 14
	OrdersDaysAgo = OZON.OrdersDaysAgo
	spreadsheetId = "1Rljs-bxCQCP0DnDfqRGw0SKm1OjRXi2EdJrW5j3M5ts"
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
		spreadsheetId: spreadsheetId,
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
