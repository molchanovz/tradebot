package yandex

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tradebot/pkg/client/yandex"
	"tradebot/pkg/tradeplus"
)

// OrdersSheetPrefix — листы ежедневного отчёта Яндекс Маркета: «Заказы YM-<число>».
const OrdersSheetPrefix = "Заказы YM-"

type OrdersManager struct {
	yandexCampaignIDFBO, yandexCampaignIDFBS, token string
}

func NewOrdersManager(yandexCampaignIDFBO, yandexCampaignIDFBS, token string) OrdersManager {
	manager := OrdersManager{yandexCampaignIDFBO, yandexCampaignIDFBS, token}
	return manager
}

// NewOrdersWriter — запись отчётов Яндекс Маркета в таблицу заказов: FBO- и
// FBS-кампании в одной секции.
func NewOrdersWriter(cabinets tradeplus.Cabinets) (tradeplus.OrdersWriter, error) {
	if len(cabinets) == 0 {
		return tradeplus.OrdersWriter{}, errors.New("нет кабинетов Яндекс Маркета")
	}
	return tradeplus.NewOrdersWriter(cabinets.OrdersSpreadsheetID(), OrdersSheetPrefix, 0, NewService(cabinets...).GetOrdersAndReturnsManager()), nil
}

// OrdersReport — заказы FBO и FBS за сутки day.
func (m OrdersManager) OrdersReport(ctx context.Context, day time.Time) (tradeplus.OrdersReport, error) {
	postingsWithCountFBO, err := m.ordersMap(ctx, m.yandexCampaignIDFBO, day)
	if err != nil {
		return tradeplus.OrdersReport{}, fmt.Errorf("fbo: %w", err)
	}

	postingsWithCountFBS, err := m.ordersMap(ctx, m.yandexCampaignIDFBS, day)
	if err != nil {
		return tradeplus.OrdersReport{}, fmt.Errorf("fbs: %w", err)
	}

	////Запись возвратов
	//returnsWithCount := returnsMap(apiKey)
	//writeRange = sheetsName + "!G2:H100"
	//colName = "Возвраты"
	//err = writeData(writeRange, colName, returnsWithCount)
	//if err != nil {
	//	return err
	//}

	return tradeplus.OrdersReport{Day: day, Sections: []tradeplus.OrdersSection{{Blocks: []tradeplus.OrdersBlock{
		{Title: "Заказы FBO", Counts: postingsWithCountFBO},
		{Title: "Заказы FBS", Counts: postingsWithCountFBS},
	}}}}, nil
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

func (m OrdersManager) ordersMap(ctx context.Context, yandexCampaignID string, day time.Time) (map[string]int, error) {
	postingsWithCountALL := make(map[string]int)
	if yandexCampaignID == "" {
		return nil, errors.New("нет кабинета с ID кампании")
	}
	ordersFbo, err := yandex.GetOrdersStats(ctx, yandexCampaignID, m.token, day)
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
