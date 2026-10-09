package yandex

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// GetOrdersStats возвращает заказы кампании, оформленные в день date, со всех страниц ответа.
func GetOrdersStats(ctx context.Context, campaignID, apiKey string, date time.Time) (OrdersFbo, error) {
	var orders OrdersFbo
	var pageToken string
	seen := make(map[int64]bool)
	for {
		jsonString, err := getOrders(ctx, campaignID, apiKey, date, pageToken)
		if err != nil {
			return orders, err
		}
		var page OrdersFbo
		if err := json.Unmarshal([]byte(jsonString), &page); err != nil {
			return orders, err
		}
		orders.Status = page.Status
		// страницы не должны пересекаться, но один заказ не считаем дважды
		for _, o := range page.Result.Orders {
			if !seen[o.ID] {
				seen[o.ID] = true
				orders.Result.Orders = append(orders.Result.Orders, o)
			}
		}

		next := page.Result.Paging.NextPageToken
		if next == "" || next == pageToken {
			return orders, nil
		}
		pageToken = next
	}
}

// func ordersFBS(apiKey string, daysAgo int) OrdersListFBS {
//	var posting OrdersListFBS
//	jsonString := API.OrdersFBS(apiKey, daysAgo)
//	err := json.Unmarshal([]byte(jsonString), &posting)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return posting
// }
//
// func salesAndReturns(apiKey string, daysAgo int) SalesAndReturns {
//	var sales SalesAndReturns
//	jsonString := API.ApiSalesAndReturns(apiKey, daysAgo)
//	err := json.Unmarshal([]byte(jsonString), &sales)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return sales
// }

// func postingStatus(apiKey string, postingId int) string {
//	var postingStatuses OrdersWithStatuses
//	jsonString := API.OrdersFBS_status(apiKey, postingId)
//	err := json.Unmarshal([]byte(jsonString), &postingStatuses)
//	if err != nil {
//		log.Fatalf("Error decoding JSON: %v", err)
//	}
//	return postingStatuses.Orders[0].WbStatus
// }

// GetOrdersIds Получение id всех заказов
func GetOrdersIds(token, supplyID string) ([]int64, error) {
	var shipment Shipment
	jsonString, err := ShipmentInfo(token, supplyID)
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
func GetOrder(token string, orderID int64) (Order, error) {
	var order Order
	jsonString, err := OrderInfo(token, orderID)
	if err != nil {
		return order, err
	}
	err = json.Unmarshal([]byte(jsonString), &order)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	return order, nil
}
