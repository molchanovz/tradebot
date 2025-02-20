package wb_stocks_analyze

import (
	"WildberriesGo_bot/pkg/api/wb"
	"encoding/json"
	"log"
)

func StockFbo(apiKey string) []Stock {
	var stocks []Stock
	stocksString := wb.StocksFbo(apiKey)
	err := json.Unmarshal([]byte(stocksString), &stocks)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return stocks
}
