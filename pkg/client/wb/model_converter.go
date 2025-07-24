package wb

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
)

func GetOrdersFbs(wildberriesKey, supplyID string) ([]OrderWB, error) {
	jsonString, err := getOrdersBySupplyID(wildberriesKey, supplyID)
	if err != nil {
		return nil, err
	}
	var orders Orders
	err = json.Unmarshal([]byte(jsonString), &orders)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %w", err)
	}
	sortOrdersByArticle(orders.Orders)
	return orders.Orders, nil
}

func GetStickersFbs(wildberriesKey string, orderID int) (StickerWB, error) {
	var stickers StickerWB
	jsonString, err := getCodesByOrderID(wildberriesKey, orderID)
	if err != nil {
		return stickers, err
	}

	err = json.Unmarshal([]byte(jsonString), &stickers)
	if err != nil {
		return stickers, fmt.Errorf("error decoding JSON: %w", err)
	}
	return stickers, nil
}

func sortOrdersByArticle(orders []OrderWB) {
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].Article < orders[j].Article
	})
}

func GetAllOrders(apiKey string, daysAgo, flag int) OrdersListALL {
	var posting OrdersListALL
	jsonString, _ := apiOrdersALL(apiKey, daysAgo, flag)
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func GetOrdersFBS(apiKey string, daysAgo int) OrdersListFBS {
	var posting OrdersListFBS
	jsonString, _ := ordersFBS(apiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func GetSalesAndReturns(apiKey string, daysAgo int) SalesReturns {
	var sales SalesReturns
	jsonString, _ := apiSalesAndReturns(apiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &sales)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return sales
}

func GetPostingStatus(apiKey string, postingID int) string {
	var postingStatuses OrdersWithStatuses
	jsonString, _ := ordersFBSStatus(apiKey, postingID)
	err := json.Unmarshal([]byte(jsonString), &postingStatuses)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return postingStatuses.Orders[0].WbStatus
}

func GetStockFbo(apiKey string) ([]Stock, error) {
	var stocks []Stock
	stocksString, err := stocksFbo(apiKey)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(stocksString), &stocks)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %w", err)
	}
	return stocks, nil
}
