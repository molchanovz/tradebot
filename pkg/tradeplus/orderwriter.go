package tradeplus

import (
	"tradebot/pkg/client/googlesheet"
)

type GoogleSheetWriter interface {
	Write() error
}

type OrderManager struct {
	DaysAgo       int
	SpreadsheetID string
	GoogleService googlesheet.SheetsService
}

func NewOrdersManager(spreadsheetID string, daysAgo int) OrderManager {
	ordersManager := OrderManager{
		DaysAgo:       daysAgo,
		SpreadsheetID: spreadsheetID,
		GoogleService: googlesheet.NewSheetsService("pkg/client/googlesheet/token.json", "pkg/client/googlesheet/credentials.json"),
	}

	return ordersManager
}
