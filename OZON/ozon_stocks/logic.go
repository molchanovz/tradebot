package ozon_stocks

import (
	"WildberriesGo_bot/OZON/API"
	"time"
)

func GetPostings(ClientId, OzonKey string, daysAgo int) map[string]map[string]int {
	since := time.Now().AddDate(0, 0, daysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, 0).Format("2006-01-02") + "T21:00:00.000Z"

	postingsListFbs := API.PostingsListFbs(ClientId, OzonKey, since, to)
	postingsListFbo := API.PostingsListFbo(ClientId, OzonKey, since, to)

	postingsMap := make(map[string]map[string]int)

	for _, order := range postingsListFbs.Result.PostingsFBS {
		if _, exists := postingsMap[order.FinancialData.ClusterTo]; !exists {
			postingsMap[order.FinancialData.ClusterTo] = make(map[string]int)
		}
		for _, product := range order.Products {
			postingsMap[order.FinancialData.ClusterTo][product.OfferId] += product.Quantity
		}
	}

	for _, order := range postingsListFbo.Result {
		if _, exists := postingsMap[order.FinancialData.ClusterTo]; !exists {
			postingsMap[order.FinancialData.ClusterTo] = make(map[string]int)
		}
		for _, product := range order.Products {
			postingsMap[order.FinancialData.ClusterTo][product.OfferId] += product.Quantity
		}
	}
	return postingsMap
}
func GetStocks(ClientId, OzonKey string) map[string]map[string]int {
	stocksList := API.Stocks(ClientId, OzonKey)

	clusters := API.Clusters(ClientId, OzonKey)

	clustersMap := make(map[string]string)

	//Разбивка складов по кластерам в map
	for _, cluster := range clusters.Clusters {
		for _, logisticClusters := range cluster.LogisticClusters {
			for _, warehouse := range logisticClusters.Warehouses {
				if _, exists := clustersMap[warehouse.Name]; !exists {
					clustersMap[warehouse.Name] = cluster.Name
				}
			}
		}
	}

	stocksMap := make(map[string]map[string]int)

	for _, stock := range stocksList.Result.Rows {
		cluster := clustersMap[stock.WarehouseName]
		if _, exists := stocksMap[cluster]; !exists {
			stocksMap[cluster] = make(map[string]int)
		}
		if _, exists := stocksMap[cluster][stock.ItemCode]; exists {
			stocksMap[cluster][stock.ItemCode] += stock.FreeToSellAmount
		} else {
			stocksMap[cluster][stock.ItemCode] = stock.FreeToSellAmount
		}

	}
	return stocksMap
}
