package wb

import (
	"encoding/json"
)

func GetOrdersFbs(wildberriesKey, supplyID string) ([]OrderWB, error) {
	var orders Orders
	jsonString, err := getOrdersBySupplyID(wildberriesKey, supplyID)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &orders)
	if err != nil {
		return nil, err
	}

	sortOrdersByArticle(orders.Orders)
	return orders.Orders, nil
}

func GetStickersFbs(wildberriesKey string, orderID int) (StickerWB, error) {
	var stickers StickerWB
	jsonString, err := getCodesByOrderID(wildberriesKey, orderID)
	if err != nil || jsonString == "" {
		return stickers, err
	}

	err = json.Unmarshal([]byte(jsonString), &stickers)
	return stickers, err
}

func GetAllOrders(apiKey string, daysAgo, flag int) (OrdersListALL, error) {
	var posting OrdersListALL
	jsonString, err := apiOrdersALL(apiKey, daysAgo, flag)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &posting)
	return posting, err
}

func GetOrdersFBS(apiKey string, daysAgo int) (*OrdersListFBS, error) {
	var posting OrdersListFBS
	jsonString, err := ordersFBS(apiKey, daysAgo)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &posting)
	return &posting, err
}

func GetSalesAndReturns(apiKey string, daysAgo int) (SalesReturns, error) {
	var sales SalesReturns
	jsonString, err := apiSalesAndReturns(apiKey, daysAgo)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &sales)
	return sales, err
}

func GetPostingStatus(apiKey string, postingID int) (string, error) {
	var postingStatuses OrdersWithStatuses
	jsonString, err := ordersFBSStatus(apiKey, postingID)
	if err != nil || jsonString == "" {
		return "", err
	}

	err = json.Unmarshal([]byte(jsonString), &postingStatuses)
	if err != nil {
		return "", err
	}

	return postingStatuses.Orders[0].WbStatus, nil
}

func GetStockFbo(apiKey string) ([]Stock, error) {
	var stocks []Stock
	jsonString, err := stocksFbo(apiKey)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &stocks)
	return stocks, nil
}
