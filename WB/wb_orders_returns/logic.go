package wb_orders_returns

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var daysAgo = 1
var spreadsheetId = "1f22DozWsJJ2L6VanQXHk81dwzRjTY2RDLEN-C4CBJB0"

func WriteToGoogleSheets(ApiKey string) error {
	date := time.Now().AddDate(0, 0, -daysAgo)
	sheetsName := "Заказы WB-" + strconv.Itoa(date.Day())

	var values [][]interface{}

	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	writeRange := sheetsName + "!A1"
	err := write(spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	//Запись ALL заказов
	postingsWithCountFBO := ordersMapALL(ApiKey)
	writeRange = sheetsName + "!A2:B100"
	colName := "Заказы FBO+FBS"
	err = writeData(writeRange, colName, postingsWithCountFBO)
	if err != nil {
		return err
	}

	//Запись FBS заказов
	postingsWithCountFBS := ordersMapFBS(ApiKey)
	writeRange = sheetsName + "!D2:E100"
	colName = "Заказы FBS"
	err = writeData(writeRange, colName, postingsWithCountFBS)
	if err != nil {
		return err
	}

	//Запись возвратов
	returnsWithCount := returnsMap(ApiKey)
	writeRange = sheetsName + "!G2:H100"
	colName = "Возвраты"
	err = writeData(writeRange, colName, returnsWithCount)
	if err != nil {
		return err
	}

	if len(postingsWithCountFBS) > len(postingsWithCountFBO) {
		return fmt.Errorf("ФБС заказов больше, чем ALL! All:%d, FBS:%d", len(postingsWithCountFBO), len(postingsWithCountFBS))
	}

	return nil
}

func writeData(writeRange, colName string, data map[string]int) error {
	var values [][]interface{}

	values = append(values, []interface{}{colName})
	for article, count := range data {
		values = append(values, []interface{}{article, count})
	}

	err := write(spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}
	return nil
}

func ordersMapFBS(ApiKey string) map[string]int {
	postingsWithCountFBS := make(map[string]int)
	postingsList := ordersFBS(ApiKey, daysAgo)

	isOrderCanceled := func(status string) bool {

		var statuses = map[string]struct{}{
			"canceled":           {},
			"canceled_by_client": {},
			"declined_by_client": {},
		}

		if _, isCancel := statuses[status]; isCancel {
			return true
		}
		return false
	}

	for _, posting := range postingsList.OrdersFBS {
		status := postingStatus(ApiKey, posting.Id)
		if !isOrderCanceled(status) {
			postingsWithCountFBS[posting.Article] += 1
		}
	}
	return postingsWithCountFBS
}
func ordersMapALL(apiKey string) map[string]int {
	postingsWithCountALL := make(map[string]int)
	postingsList := allOrders(apiKey, daysAgo)
	for _, posting := range postingsList {
		if posting.OrderType == "Клиентский" && posting.IsCancel == false {
			postingsWithCountALL[posting.SupplierArticle]++
		} else {
			fmt.Println("Пропавший заказ ", posting.SupplierArticle, " ", posting.IsCancel)
		}
	}
	return postingsWithCountALL
}
func returnsMap(apiKey string) map[string]int {
	returnsWithCount := make(map[string]int)
	returnsList := salesAndReturns(apiKey, daysAgo)
	for _, someReturn := range returnsList {
		if strings.HasPrefix(someReturn.SaleID, "R") {
			returnsWithCount[someReturn.SupplierArticle]++
		}
	}
	return returnsWithCount
}

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
