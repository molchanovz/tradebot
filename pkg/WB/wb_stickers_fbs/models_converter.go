package wb_stickers_fbs

import (
	"WildberriesGo_bot/pkg/api/wb"
	"encoding/json"
	"fmt"
	"log"
	"sort"
)

func GetOrdersFbs(wildberriesKey, supplyId string) ([]OrderWB, error) {
	jsonString, err := wb.GetOrdersBySupplyId(wildberriesKey, supplyId)
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

func GetStickersFbs(wildberriesKey string, orderId int) StickerWB {
	jsonString := wb.GetCodesByOrderId(wildberriesKey, orderId)
	var stickers StickerWB
	err := json.Unmarshal([]byte(jsonString), &stickers)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	return stickers
}

func sortOrdersByArticle(orders []OrderWB) {
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].Article < orders[j].Article
	})
}
