package ozon

import (
	"strconv"
	"time"

	"tradebot/pkg/client/ozon"
)

type Period struct {
	since string
	to    string
}

type AnalyzeManager struct {
	daysAgo         int
	clientID, token string
	//googleService   google.SheetsService
}

func NewAnalyzeManager(clientID, token string, daysAgo int) AnalyzeManager {
	return AnalyzeManager{
		clientID: clientID,
		token:    token,
		daysAgo:  daysAgo,
		//googleService: google.NewSheetsService("token.json", "credentials.json"),
	}
}

func (m AnalyzeManager) GetPostings() map[string]map[string]map[string]int {

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
			postingsListFbs, _ := ozon.PostingsListFbs(m.clientID, m.token, period.since, period.to, offset, "")

			for _, order := range postingsListFbs.Result.PostingsFBS {
				cluster := order.FinancialData.ClusterTo
				if _, exists := postingsMap[cluster]; !exists {
					postingsMap[cluster] = make(map[string]map[string]int)
				}

				for _, product := range order.Products {
					if _, exists := postingsMap[cluster][product.OfferID]; !exists {
						postingsMap[cluster][product.OfferID] = make(map[string]int)
					}

					postingsMap[cluster][product.OfferID][date] += product.Quantity
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
			postingsListFbo := ozon.PostingsListFbo(m.clientID, m.token, period.since, period.to, offset)

			for _, order := range postingsListFbo.Result {
				cluster := order.FinancialData.ClusterTo
				if _, exists := postingsMap[cluster]; !exists {
					postingsMap[cluster] = make(map[string]map[string]int)
				}

				for _, product := range order.Products {
					if _, exists := postingsMap[cluster][product.OfferID]; !exists {
						postingsMap[cluster][product.OfferID] = make(map[string]int)
					}

					postingsMap[cluster][product.OfferID][date] += product.Quantity
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

func (m AnalyzeManager) GetStocks() map[string]map[string]CustomStocks {
	products := ozon.ProductsWithAttributes(m.clientID, m.token)

	skus := make([]string, 0, len(products.Result))

	for _, item := range products.Result {
		skus = append(skus, strconv.Itoa(item.Sku))
	}

	maxSkus := 100
	stocksMap := make(map[string]map[string]CustomStocks)

	for i := 0; i < len(skus); i += maxSkus {
		end := i + maxSkus
		if end > len(skus) {
			end = len(skus)
		}
		chunk := skus[i:end]

		stocksList := ozon.StocksAnalytics(m.clientID, m.token, chunk)

		for _, item := range stocksList.Items {
			if _, exists := stocksMap[item.ClusterName]; !exists {
				stocksMap[item.ClusterName] = make(map[string]CustomStocks)
			}
			if stock, exists := stocksMap[item.ClusterName][item.OfferID]; exists {
				//TODO Добавить все остатки
				stock.AvailableStockCount += item.AvailableStockCount
				stock.RequestedStockCount += item.RequestedStockCount
				stock.TransitStockCount += item.TransitStockCount
				stocksMap[item.ClusterName][item.OfferID] = stock
			} else {
				stock.AvailableStockCount = item.AvailableStockCount
				stock.RequestedStockCount = item.RequestedStockCount
				stock.TransitStockCount = item.TransitStockCount
				stocksMap[item.ClusterName][item.OfferID] = stock
			}
		}
	}

	return stocksMap
}

func (m AnalyzeManager) GetClusters() ozon.ClustersList {
	return ozon.Clusters(m.clientID, m.token)
}
