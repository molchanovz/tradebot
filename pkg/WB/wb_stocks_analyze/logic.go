package wb_stocks_analyze

import (
	"tradebot/pkg/api/wb"
)

var warehousesMap = map[string]string{
	"Санкт-Петербург Уткина Заводь":  "Северо-Западный федеральный округ",
	"Екатеринбург - Испытателей 14г": "Уральский федеральный округ",
	"Невинномысск":                   "Северо-Кавказский федеральный округ",
	"Новосибирск":                    "Сибирский федеральный округ",
	"Краснодар":                      "Южный федеральный округ",
	"Волгоград":                      "Южный федеральный округ",
	"Подольск":                       "Центральный федеральный округ",
	"Подольск 3":                     "Центральный федеральный округ",
	"Подольск 4":                     "Центральный федеральный округ",
	"Белые столбы":                   "Центральный федеральный округ",
	"Коледино":                       "Центральный федеральный округ",
	"Белая дача":                     "Центральный федеральный округ",
	"Электросталь":                   "Центральный федеральный округ",
	"Чашниково":                      "Центральный федеральный округ",
	"Тула":                           "Центральный федеральный округ",
	"Котовск":                        "Центральный федеральный округ",
	"Владимир":                       "Центральный федеральный округ",
	"Казань":                         "Приволжский федеральный округ",
	"Хабаровск":                      "Дальневосточный федеральный округ",
	"Рязань (Тюшевское)":             "Центральный федеральный округ",
}

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

func GetStocks(apiKey string) (map[string]map[string]int, map[string]int, error) {
	stocks, err := wb.GetStockFbo(apiKey)
	if err != nil {
		return nil, nil, err
	}

	stocksMap := make(map[string]map[string]int)

	lostWarehouses := make(map[string]int)

	for _, stock := range stocks {
		if federalRegion, ok := warehousesMap[stock.WarehouseName]; !ok && stock.Quantity > 0 {
			// Ищем потерянные склады с округах
			lostWarehouses[stock.WarehouseName] = 1
		} else {
			stock.WarehouseName = federalRegion
		}

		if _, exists := stocksMap[stock.WarehouseName]; !exists {
			stocksMap[stock.WarehouseName] = make(map[string]int)
		}

		stocksMap[stock.WarehouseName][stock.SupplierArticle] += stock.Quantity
	}
	return stocksMap, lostWarehouses, nil
}
