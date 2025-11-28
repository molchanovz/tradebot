package yandex

import (
	"strconv"
	"time"

	"tradebot/pkg/client/yandex"
	"tradebot/pkg/tradeplus"
)

type OrdersManager struct {
	tradeplus.OrderManager
	yandexCampaignIDFBO, yandexCampaignIDFBS, token string
}

func NewOrdersManager(yandexCampaignIDFBO, yandexCampaignIDFBS, token, spreadsheetID string) OrdersManager {
	manager := OrdersManager{tradeplus.NewOrdersManager(spreadsheetID), yandexCampaignIDFBO, yandexCampaignIDFBS, token}
	return manager
}

func (m OrdersManager) Write() error {
	date := time.Now().AddDate(0, 0, -m.DaysAgo)
	sheetsName := "Заказы YM-" + strconv.Itoa(date.Day())

	var values [][]interface{}
	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	writeRange := sheetsName + "!A1"
	err := m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return err
	}

	// Запись FBO заказов
	postingsWithCountFBO, err := m.ordersMap(m.yandexCampaignIDFBO)
	if err != nil {
		return err
	}
	writeRange = sheetsName + "!A2:B100"
	colName := "Заказы FBO"
	values = [][]interface{}{}
	values = append(values, []interface{}{colName})
	for article, count := range postingsWithCountFBO {
		values = append(values, []interface{}{article, count})
	}
	err = m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return err
	}

	//Запись FBS заказов
	postingsWithCountFBS, err := m.ordersMap(m.yandexCampaignIDFBS)
	if err != nil {
		return err
	}
	writeRange = sheetsName + "!D2:E100"
	colName = "Заказы FBS"
	values = [][]interface{}{}
	values = append(values, []interface{}{colName})
	for article, count := range postingsWithCountFBS {
		values = append(values, []interface{}{article, count})
	}
	err = m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return err
	}

	////Запись возвратов
	//returnsWithCount := returnsMap(apiKey)
	//writeRange = sheetsName + "!G2:H100"
	//colName = "Возвраты"
	//err = writeData(writeRange, colName, returnsWithCount)
	//if err != nil {
	//	return err
	//}

	return nil
}

//func ordersMapFBS(apiKey string) map[string]int {
//	postingsWithCountFBS := make(map[string]int)
//	postingsList := ordersFBS(apiKey, DaysAgo)
//
//	isOrderCanceled := func(status string) bool {
//
//		var statuses = map[string]struct{}{
//			"canceled":           {},
//			"canceled_by_client": {},
//			"declined_by_client": {},
//		}
//
//		if _, isCancel := statuses[status]; isCancel {
//			return true
//		}
//		return false
//	}
//
//	for _, posting := range postingsList.OrdersFBS {
//		status := postingStatus(apiKey, posting.Id)
//		if !isOrderCanceled(status) {
//			postingsWithCountFBS[posting.Article] += 1
//		}
//	}
//	return postingsWithCountFBS
//}

func (m OrdersManager) ordersMap(yandexCampaignID string) (map[string]int, error) {
	postingsWithCountALL := make(map[string]int)
	ordersFbo, err := yandex.GetOrdersFbo(yandexCampaignID, m.token, m.DaysAgo)
	if err != nil {
		return postingsWithCountALL, err
	}
	for _, order := range ordersFbo.Result.Orders {
		for _, items := range order.Items {
			postingsWithCountALL[items.ShopSku] += items.Count
		}
	}
	return postingsWithCountALL, nil
}

//func returnsMap(apiKey string) map[string]int {
//	returnsWithCount := make(map[string]int)
//	returnsList := salesAndReturns(apiKey, DaysAgo)
//	for _, someReturn := range returnsList {
//		if strings.HasPrefix(someReturn.SaleID, "R") {
//			returnsWithCount[someReturn.SupplierArticle]++
//		}
//	}
//	return returnsWithCount
//}

//func initEnv(path, name string) (string, error) {
//	err := godotenv.Load(path)
//	if err != nil {
//		log.Printf("Ошибка загрузки файла %s: %v\n", path, err)
//		return "", fmt.Errorf("ошибка загрузки файла " + path)
//	}
//	// Получаем значения переменных среды
//	env := os.Getenv(name)
//
//	if env == "" {
//		return "", fmt.Errorf("переменная среды " + name + " не установлена")
//	}
//	return env, err
//}

//func getUnix(date time.Time) int64 {
//	nowStr := fmt.Sprint(date.Format("2006-01-02"), "T21:00:00")
//	t, _ := time.Parse("2006-01-02T15:04:05", nowStr)
//	return t.Unix()
//}
