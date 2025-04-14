package stocks_analyzer

import (
	"WildberriesGo_bot/pkg/api/ozon"
	"WildberriesGo_bot/pkg/google"
	"time"
)

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

func (m OzonManager) GetPostings() map[string]map[string]int {
	since := time.Now().AddDate(0, 0, m.daysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, 0).Format("2006-01-02") + "T21:00:00.000Z"

	postingsListFbs := ozon.PostingsListFbs(m.clientId, m.token, since, to, 0, "")
	postingsListFbo := ozon.PostingsListFbo(m.clientId, m.token, since, to, 0)

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
func (m OzonManager) GetStocks() map[string]map[string]int {
	stocksList := ozon.Stocks(m.clientId, m.token)

	clusters := ozon.Clusters(m.clientId, m.token)

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
			stocksMap[cluster][stock.ItemCode] += stock.FreeToSellAmount + stock.PromisedAmount
		} else {
			stocksMap[cluster][stock.ItemCode] = stock.FreeToSellAmount + stock.PromisedAmount
		}

	}
	return stocksMap
}
func (m OzonManager) GetClusters() ozon.ClustersList {
	return ozon.Clusters(m.clientId, m.token)
}
