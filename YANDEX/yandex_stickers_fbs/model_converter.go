package yandex_stickers_fbs

import (
	"WildberriesGo_bot/YANDEX/API"
	"encoding/json"
	"log"
)

// Получение id всех заказов
func GetOrdersIds(token, supplyId string) ([]int64, error) {
	var shipment Shipment
	jsonString, err := API.ShipmentInfo(token, supplyId)
	if err != nil {
		return []int64{}, err
	}
	err = json.Unmarshal([]byte(jsonString), &shipment)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return shipment.Result.OrderIds, nil
}

// Получение id всех заказов
func GetOrder(token string, orderId int64) (Order, error) {
	var order Order
	jsonString, err := API.OrderInfo(token, orderId)
	if err != nil {
		return order, err
	}
	err = json.Unmarshal([]byte(jsonString), &order)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	return order, nil
}
