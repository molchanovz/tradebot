package OrdersAndReturns

import (
	"WildberriesGo_bot/pkg/api/yandex"
	"WildberriesGo_bot/pkg/googleService"
	"strconv"
	"time"
)

type Manager struct {
	token, spreadsheetId string
	daysAgo              int
	googleService        googleService.GoogleService
}

func NewManager(token, spreadsheetId string, daysAgo int) *Manager {
	return &Manager{
		spreadsheetId: spreadsheetId,
		daysAgo:       daysAgo,
		token:         token,
		googleService: googleService.NewGoogleService("token.json", "credentials.json"),
	}
}

func (m Manager) WriteToGoogleSheets() error {
	date := time.Now().AddDate(0, 0, -m.daysAgo)
	sheetsName := "Заказы YM-" + strconv.Itoa(date.Day())

	var values [][]interface{}

	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	writeRange := sheetsName + "!A1"
	err := m.googleService.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	//Запись FBO заказов
	postingsWithCountFBO, _ := m.ordersMapFBO()
	writeRange = sheetsName + "!A2:B100"
	colName := "Заказы FBO"
	values = [][]interface{}{}
	values = append(values, []interface{}{colName})
	for article, count := range postingsWithCountFBO {
		values = append(values, []interface{}{article, count})
	}
	err = m.googleService.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	////Запись FBS заказов
	//postingsWithCountFBS := ordersMapFBS(ApiKey)
	//writeRange = sheetsName + "!D2:E100"
	//colName = "Заказы FBS"
	//err = writeData(writeRange, colName, postingsWithCountFBS)
	//if err != nil {
	//	return err
	//}

	////Запись возвратов
	//returnsWithCount := returnsMap(ApiKey)
	//writeRange = sheetsName + "!G2:H100"
	//colName = "Возвраты"
	//err = writeData(writeRange, colName, returnsWithCount)
	//if err != nil {
	//	return err
	//}

	return nil
}

//func ordersMapFBS(ApiKey string) map[string]int {
//	postingsWithCountFBS := make(map[string]int)
//	postingsList := ordersFBS(ApiKey, DaysAgo)
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
//		status := postingStatus(ApiKey, posting.Id)
//		if !isOrderCanceled(status) {
//			postingsWithCountFBS[posting.Article] += 1
//		}
//	}
//	return postingsWithCountFBS
//}

func (m Manager) ordersMapFBO() (map[string]int, error) {
	postingsWithCountALL := make(map[string]int)
	ordersFbo, err := yandex.GetOrdersFbo(m.token, m.daysAgo)
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
