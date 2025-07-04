package OZON

import (
	"math"
	"strconv"
	"time"
	"tradebot/pkg/marketplaces/OZON/api"
	"tradebot/pkg/ordersWriter"
)

type OrdersManager struct {
	ordersWriter.OrdersManager
	clientId, token string
}

func NewOrdersManager(clientId, token, spreadsheetId string, daysAgo int) OrdersManager {
	manager := OrdersManager{ordersWriter.NewOrdersManager(spreadsheetId, daysAgo), clientId, token}
	return manager
}

// WriteToGoogleSheets Заполнение гугл таблицы с id = spreadsheetId
func (m OrdersManager) WriteToGoogleSheets(titleRange, fbsRange, fboRange, returnsRange string) (int, error) {
	date := time.Now().AddDate(0, 0, -m.DaysAgo)
	sheetsName := "Заказы OZON-" + strconv.Itoa(date.Day())

	maxValuesCount := math.MinInt

	var values [][]interface{}
	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	if maxValuesCount < len(values) {
		maxValuesCount = len(values)
	}

	writeRange := sheetsName + titleRange

	err := m.GoogleService.Write(m.SpreadsheetId, writeRange, values)
	if err != nil {
		return 0, err
	}

	//Заполнение заказов FBS в writeRange
	postingsWithCountFBS := m.getPostingsMapFBS(m.clientId, m.token)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBS"})
	for article, count := range postingsWithCountFBS {
		values = append(values, []interface{}{article, count})
	}

	if maxValuesCount < len(values) {
		maxValuesCount = len(values)
	}

	writeRange = sheetsName + fbsRange
	err = m.GoogleService.Write(m.SpreadsheetId, writeRange, values)
	if err != nil {
		return 0, err
	}

	postingsWithCountFBO := m.getPostingsMapFBO(m.clientId, m.token)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBO"})
	for article, count := range postingsWithCountFBO {
		values = append(values, []interface{}{article, count})
	}

	if maxValuesCount < len(values) {
		maxValuesCount = len(values)
	}

	writeRange = sheetsName + fboRange
	err = m.GoogleService.Write(m.SpreadsheetId, writeRange, values)
	if err != nil {
		return 0, err
	}

	since := time.Now().AddDate(0, 0, m.DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	//Заполнение возвратов
	returnsWithCount, err := m.getReturnsMap(m.clientId, m.token, since, to)
	if err != nil {
		return 0, err
	}

	values = [][]interface{}{}
	values = append(values, []interface{}{"Возвраты"})
	for article, count := range returnsWithCount {
		values = append(values, []interface{}{article, count})
	}

	if maxValuesCount < len(values) {
		maxValuesCount = len(values)
	}

	writeRange = sheetsName + returnsRange
	err = m.GoogleService.Write(m.SpreadsheetId, writeRange, values)
	if err != nil {
		return 0, err
	}

	return maxValuesCount, nil
}

func (m OrdersManager) getPostingsMapFBS(clientId, token string) map[string]int {
	postingsWithCountFBS := make(map[string]int)
	since := time.Now().AddDate(0, 0, m.DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	potingsListFbs, _ := api.PostingsListFbs(clientId, token, since, to, 0, "")
	for _, posting := range potingsListFbs.Result.PostingsFBS {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBS[product.OfferId] += product.Quantity
			}
		}
	}
	return postingsWithCountFBS
}
func (m OrdersManager) getPostingsMapFBO(clientId, token string) map[string]int {
	postingsWithCountFBO := make(map[string]int)
	since := time.Now().AddDate(0, 0, m.DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	postings_list_fbo := api.PostingsListFbo(clientId, token, since, to, 0)
	for _, posting := range postings_list_fbo.Result {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBO[product.OfferId] += product.Quantity
			}
		}
	}
	return postingsWithCountFBO
}
func (m OrdersManager) getReturnsMap(clientId, token, since, to string) (map[string]int, error) {
	LastID := 0
	hasNext := true
	returnsWithCount := make(map[string]int)
	/*
		Лимит у запроса 1000, но нам нужны все возвраты,
		поэтому делаем цикл с LastID и добавляем в срез returnsFBO
	*/
	for hasNext {
		returns, err := api.ReturnsList(clientId, token, LastID, since, to)
		if err != nil {
			return returnsWithCount, err
		}
		for _, value := range returns.Returns {
			if value.Visual.Status.SysName == "ReturnedToOZON" {
				returnsWithCount[value.Product.OfferId] += value.Product.Quantity
			}
			LastID = value.Id
		}
		hasNext = returns.HasNext
	}
	return returnsWithCount, nil
}
