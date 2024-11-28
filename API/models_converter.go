package API

import (
	"encoding/json"
	"log"
)

func StockFbo(apiKey string) []Stock {
	var stocks []Stock
	stocksString := stocksFbo(apiKey)
	err := json.Unmarshal([]byte(stocksString), &stocks)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return stocks
}
