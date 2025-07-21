package tradeplus

import (
	"tradebot/pkg/client/googlesheet"
)

type GoogleSheetWriter interface {
	Write() error
}

type OrderManager struct {
	manager       GoogleSheetWriter
	DaysAgo       int
	SpreadsheetId string
	GoogleService googlesheet.SheetsService
}

func NewOrdersManager(spreadsheetId string, daysAgo int) OrderManager {
	ordersManager := OrderManager{
		DaysAgo:       daysAgo,
		SpreadsheetId: spreadsheetId,
		GoogleService: googlesheet.NewSheetsService("pkg/client/googlesheet/token.json", "pkg/client/googlesheet/credentials.json"),
	}

	return ordersManager
}
