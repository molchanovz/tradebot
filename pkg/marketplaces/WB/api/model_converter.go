package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
)

func GetOrdersFbs(wildberriesKey, supplyId string) ([]OrderWB, error) {
	jsonString, err := getOrdersBySupplyId(wildberriesKey, supplyId)
	if err != nil {
		return nil, err
	}
	var orders Orders
	err = json.Unmarshal([]byte(jsonString), &orders)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}
	sortOrdersByArticle(orders.Orders)
	return orders.Orders, nil
}

func GetStickersFbs(wildberriesKey string, orderId int) (StickerWB, error) {
	var stickers StickerWB
	jsonString, err := getCodesByOrderId(wildberriesKey, orderId)
	if err != nil {
		return stickers, err
	}

	err = json.Unmarshal([]byte(jsonString), &stickers)
	if err != nil {
		return stickers, errors.New(fmt.Sprintf("error decoding JSON: %v", err))
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

func GetOrdersFBS(ApiKey string, daysAgo int) OrdersListFBS {
	var posting OrdersListFBS
	jsonString, _ := ordersFBS(ApiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &posting)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return posting
}

func GetSalesAndReturns(ApiKey string, daysAgo int) SalesReturns {
	var sales SalesReturns
	jsonString, _ := apiSalesAndReturns(ApiKey, daysAgo)
	err := json.Unmarshal([]byte(jsonString), &sales)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return sales
}

func GetPostingStatus(ApiKey string, postingId int) string {
	var postingStatuses OrdersWithStatuses
	jsonString, _ := ordersFBSStatus(ApiKey, postingId)
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
		return nil, fmt.Errorf("Error decoding JSON: %v", err)
	}
	return stocks, nil
}
