package ozon

import (
	"log"
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
		client := ozon.NewClient(m.clientID, m.token)

		// Обработка FBS заказов
		postingsListFbs, err := client.PostingsListFbs(ozon.FbsFilter{Since: period.since, To: period.to})
		if err != nil {
			log.Printf("ozonStocks: clientID=%s FBS за %s: %v", m.clientID, date, err)
		}
		for _, order := range postingsListFbs.Result.PostingsFBS {
			addPostings(postingsMap, order.FinancialData.ClusterTo, order.Products, date)
		}

		// Обработка FBO заказов
		postingsListFbo, err := client.PostingsListFbo(period.since, period.to)
		if err != nil {
			log.Printf("ozonStocks: clientID=%s FBO за %s: %v", m.clientID, date, err)
		}
		for _, order := range postingsListFbo.Result {
			addPostings(postingsMap, order.FinancialData.ClusterTo, order.Products, date)
		}
	}

	return postingsMap
}

// addPostings добавляет количество товаров отправления в postingsMap[кластер][артикул][дата].
func addPostings(postingsMap map[string]map[string]map[string]int, cluster string, products []ozon.PostingProduct, date string) {
	if _, exists := postingsMap[cluster]; !exists {
		postingsMap[cluster] = make(map[string]map[string]int)
	}
	for _, product := range products {
		if _, exists := postingsMap[cluster][product.OfferID]; !exists {
			postingsMap[cluster][product.OfferID] = make(map[string]int)
		}
		postingsMap[cluster][product.OfferID][date] += product.Quantity
	}
}

type CustomStocks struct {
	AvailableStockCount int
	TransitStockCount   int
	RequestedStockCount int
}

func (m AnalyzeManager) GetStocks() (map[string]map[string]CustomStocks, error) {
	products, err := ozon.NewClient(m.clientID, m.token).ProductsWithAttributes()
	if err != nil {
		return nil, err
	}

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

		stocksList, _ := ozon.NewClient(m.clientID, m.token).StocksAnalytics(chunk)

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

	return stocksMap, nil
}

func (m AnalyzeManager) GetClusters() (ozon.ClustersList, error) {
	return ozon.NewClient(m.clientID, m.token).Clusters()
}
