package stocks_analyzer

import (
	"fmt"
	"time"
	"tradebot/pkg/api/ozon"
	"tradebot/pkg/google"
)

type Period struct {
	since string
	to    string
}

type OzonManager struct {
	daysAgo         int
	clientId, token string
	googleService   google.SheetsService
}

func NewManager(clientId, token string, daysAgo int) OzonManager {
	return OzonManager{
		clientId:      clientId,
		token:         token,
		daysAgo:       daysAgo,
		googleService: google.NewSheetsService("token.json", "credentials.json"),
	}
}

func (m OzonManager) GetPostings() map[string]map[string]map[string]int {

	postingsMap := make(map[string]map[string]map[string]int)
	dates := make(map[string]Period)

	for i := 0; i < m.daysAgo; i++ {

		date := time.Now().AddDate(0, 0, -i-1).Format("2006-01-02")

		newPeriod := Period{
			since: time.Now().AddDate(0, 0, -i-2).Format("2006-01-02") + "T21:00:00.000Z",
			to:    time.Now().AddDate(0, 0, -i-1).Format("2006-01-02") + "T21:00:00.000Z",
		}

		dates[date] = newPeriod
	}

	for date, period := range dates {

		// Обработка FBS заказов
		offset := 0
		limit := 1000
		for {
			postingsListFbs := ozon.PostingsListFbs(m.clientId, m.token, period.since, period.to, offset, "")

			for _, order := range postingsListFbs.Result.PostingsFBS {

				cluster := order.FinancialData.ClusterTo
				if _, exists := postingsMap[cluster]; !exists {
					postingsMap[cluster] = make(map[string]map[string]int)
				}

				for _, product := range order.Products {

					if _, exists := postingsMap[cluster][product.OfferId]; !exists {
						postingsMap[cluster][product.OfferId] = make(map[string]int)
					}

					postingsMap[cluster][product.OfferId][date] += product.Quantity
				}
			}

			if !postingsListFbs.Result.HasNext || len(postingsListFbs.Result.PostingsFBS) < limit {
				break
			}
			offset += limit
		}

		// Обработка FBO заказов
		offset = 0
		for {
			postingsListFbo := ozon.PostingsListFbo(m.clientId, m.token, period.since, period.to, offset)

			for _, order := range postingsListFbo.Result {

				cluster := order.FinancialData.ClusterTo
				if _, exists := postingsMap[cluster]; !exists {
					postingsMap[cluster] = make(map[string]map[string]int)
				}

				for _, product := range order.Products {

					if _, exists := postingsMap[cluster][product.OfferId]; !exists {
						postingsMap[cluster][product.OfferId] = make(map[string]int)
					}

					postingsMap[cluster][product.OfferId][date] += product.Quantity
				}
			}

			if len(postingsListFbo.Result) < limit {
				break
			}
			offset += limit
		}
	}

	return postingsMap
}

type CustomStocks struct {
	AvailableStockCount int
	TransitStockCount   int
	RequestedStockCount int
}

func (m OzonManager) GetStocks() map[string]map[string]CustomStocks {

	products := ozon.ProductsWithAttributes(m.clientId, m.token)

	skus := make([]string, 0, len(products.Result))

	for _, item := range products.Result {
		skus = append(skus, fmt.Sprintf("%v", item.Sku))
	}

	maxSkus := 100
	stocksMap := make(map[string]map[string]CustomStocks)

	for i := 0; i < len(skus); i += maxSkus {
		end := i + maxSkus
		if end > len(skus) {
			end = len(skus)
		}
		chunk := skus[i:end]

		stocksList := ozon.StocksAnalytics(m.clientId, m.token, chunk)

		for _, item := range stocksList.Items {
			if _, exists := stocksMap[item.ClusterName]; !exists {
				stocksMap[item.ClusterName] = make(map[string]CustomStocks)
			}
			if stock, exists := stocksMap[item.ClusterName][item.OfferId]; exists {
				//TODO Добавить все остатки
				stock.AvailableStockCount += item.AvailableStockCount
				stock.RequestedStockCount += item.RequestedStockCount
				stock.TransitStockCount += item.TransitStockCount
				stocksMap[item.ClusterName][item.OfferId] = stock
			} else {
				stock.AvailableStockCount = item.AvailableStockCount
				stock.RequestedStockCount = item.RequestedStockCount
				stock.TransitStockCount = item.TransitStockCount
				stocksMap[item.ClusterName][item.OfferId] = stock
			}
		}
	}

	return stocksMap
}

//func (m OzonManager) GetStocks() map[string]map[string]int {
//	stocksList := ozon.Stocks(m.clientId, m.token)
//
//	clusters := ozon.Clusters(m.clientId, m.token)
//
//	clustersMap := make(map[string]string)
//
//	//Разбивка складов по кластерам в map
//	for _, cluster := range clusters.Clusters {
//		for _, logisticClusters := range cluster.LogisticClusters {
//			for _, warehouse := range logisticClusters.Warehouses {
//				if _, exists := clustersMap[warehouse.Name]; !exists {
//					clustersMap[warehouse.Name] = cluster.Name
//				}
//			}
//		}
//	}
//
//	stocksMap := make(map[string]map[string]int)
//
//	for _, stock := range stocksList.Result.Rows {
//		cluster := clustersMap[stock.WarehouseName]
//		if _, exists := stocksMap[cluster]; !exists {
//			stocksMap[cluster] = make(map[string]int)
//		}
//		if _, exists := stocksMap[cluster][stock.ItemCode]; exists {
//			stocksMap[cluster][stock.ItemCode] += stock.FreeToSellAmount + stock.PromisedAmount
//		} else {
//			stocksMap[cluster][stock.ItemCode] = stock.FreeToSellAmount + stock.PromisedAmount
//		}
//
//	}
//	return stocksMap
//}

func (m OzonManager) GetClusters() ozon.ClustersList {
	return ozon.Clusters(m.clientId, m.token)
}
