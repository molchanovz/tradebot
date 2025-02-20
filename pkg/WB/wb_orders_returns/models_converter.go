package wb_orders_returns

import (
	"WildberriesGo_bot/pkg/api/wb"
	"encoding/json"
	"log"
)

func allOrders(apiKey string, daysAgo int) OrdersListALL {
	var posting OrdersListALL
	jsonString := wb.ApiOrdersALL(apiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func ordersFBS(ApiKey string, daysAgo int) OrdersListFBS {
	var posting OrdersListFBS
	jsonString := wb.OrdersFBS(ApiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func salesAndReturns(ApiKey string, daysAgo int) SalesAndReturns {
	var sales SalesAndReturns
	jsonString := wb.ApiSalesAndReturns(ApiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &sales)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return sales
}

func postingStatus(ApiKey string, postingId int) string {
	var postingStatuses OrdersWithStatuses
	jsonString := wb.OrdersFBS_status(ApiKey, postingId)
	err := json.Unmarshal([]byte(jsonString), &postingStatuses)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return postingStatuses.Orders[0].WbStatus
}
