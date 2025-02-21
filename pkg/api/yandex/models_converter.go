package yandex

import (
	"encoding/json"
	"log"
)

func GetOrdersFbo(apiKey string, daysAgo int) (OrdersFbo, error) {
	var orders OrdersFbo
	jsonString, err := getOrdersFbo(apiKey, daysAgo)
	if err != nil {
		return orders, err
	}
	err = json.Unmarshal([]byte(jsonString), &orders)
	if err != nil {
		return orders, err
	}
	return orders, nil
}

//func ordersFBS(ApiKey string, daysAgo int) OrdersListFBS {
//	var posting OrdersListFBS
//	jsonString := API.OrdersFBS(ApiKey, daysAgo)
//	err := json.Unmarshal([]byte(jsonString), &posting)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return posting
//}
//
//func salesAndReturns(ApiKey string, daysAgo int) SalesAndReturns {
//	var sales SalesAndReturns
//	jsonString := API.ApiSalesAndReturns(ApiKey, daysAgo)
//	err := json.Unmarshal([]byte(jsonString), &sales)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return sales
//}

//func postingStatus(ApiKey string, postingId int) string {
//	var postingStatuses OrdersWithStatuses
//	jsonString := API.OrdersFBS_status(ApiKey, postingId)
//	err := json.Unmarshal([]byte(jsonString), &postingStatuses)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return postingStatuses.Orders[0].WbStatus
//}

// GetOrdersIds Получение id всех заказов
func GetOrdersIds(token, supplyId string) ([]int64, error) {
	var shipment Shipment
	jsonString, err := ShipmentInfo(token, supplyId)
	if err != nil {
		return []int64{}, err
	}
	err = json.Unmarshal([]byte(jsonString), &shipment)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return shipment.Result.OrderIds, nil
}

// GetOrder Получение заказа
func GetOrder(token string, orderId int64) (Order, error) {
	var order Order
	jsonString, err := OrderInfo(token, orderId)
	if err != nil {
		return order, err
	}
	err = json.Unmarshal([]byte(jsonString), &order)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	return order, nil
}
