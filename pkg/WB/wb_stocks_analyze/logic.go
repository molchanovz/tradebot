package wb_stocks_analyze

import "WildberriesGo_bot/pkg/api/wb"

func GetOrders(apiKey string, daysAgo int) map[string]map[string]int {
	orders := wb.GetAllOrders(apiKey, daysAgo, 0)

	ordersMap := make(map[string]map[string]int)

	for _, order := range orders {
		cluster := order.OblastOkrugName
		if cluster == "" {
			cluster = order.CountryName
		}
		if _, exists := ordersMap[cluster]; !exists {
			ordersMap[cluster] = make(map[string]int)
		}
		ordersMap[cluster][order.SupplierArticle] += 1

	}
	return ordersMap
}
