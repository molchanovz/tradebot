package ordersWriter

import "tradebot/pkg/google"

type Manager interface {
	WriteToGoogleSheets() error // Теперь метод возвращает ошибку
}

type OrdersManager struct {
	manager       Manager
	DaysAgo       int
	SpreadsheetId string
	GoogleService google.SheetsService
}

func NewOrdersManager(spreadsheetId string, daysAgo int) OrdersManager {
	ordersManager := OrdersManager{
		DaysAgo:       daysAgo,
		SpreadsheetId: spreadsheetId,
		GoogleService: google.NewSheetsService("pkg/google/utils/token.json", "pkg/google/utils/credentials.json"),
	}

	return ordersManager
}
