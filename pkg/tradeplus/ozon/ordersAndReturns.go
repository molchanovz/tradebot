package ozon

import (
	"math"
	"strconv"
	"time"

	"tradebot/pkg/client/ozon"
	"tradebot/pkg/tradeplus"
)

type OrdersManager struct {
	tradeplus.OrderManager
	clientID, token string
}

func NewOrdersManager(clientID, token, spreadsheetID string, daysAgo int) OrdersManager {
	manager := OrdersManager{tradeplus.NewOrdersManager(spreadsheetID, daysAgo), clientID, token}
	return manager
}

// WriteToGoogleSheets Заполнение гугл таблицы с id = spreadsheetID
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

	err := m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return 0, err
	}

	//Заполнение заказов FBS в writeRange
	postingsWithCountFBS := m.getPostingsMapFBS(m.clientID, m.token)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBS"})
	for article, count := range postingsWithCountFBS {
		values = append(values, []interface{}{article, count})
	}

	if maxValuesCount < len(values) {
		maxValuesCount = len(values)
	}

	writeRange = sheetsName + fbsRange
	err = m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return 0, err
	}

	postingsWithCountFBO := m.getPostingsMapFBO(m.clientID, m.token)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBO"})
	for article, count := range postingsWithCountFBO {
		values = append(values, []interface{}{article, count})
	}

	if maxValuesCount < len(values) {
		maxValuesCount = len(values)
	}

	writeRange = sheetsName + fboRange
	err = m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return 0, err
	}

	since := time.Now().AddDate(0, 0, m.DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	//Заполнение возвратов
	returnsWithCount, err := m.GetReturnsMap(m.clientID, m.token, since, to)
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
	err = m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return 0, err
	}

	return maxValuesCount, nil
}

func (m OrdersManager) getPostingsMapFBS(clientID, token string) map[string]int {
	postingsWithCountFBS := make(map[string]int)
	since := time.Now().AddDate(0, 0, m.DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	potingsListFbs, _ := ozon.PostingsListFbs(clientID, token, since, to, 0, "")
	for _, posting := range potingsListFbs.Result.PostingsFBS {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBS[product.OfferID] += product.Quantity
			}
		}
	}
	return postingsWithCountFBS
}
func (m OrdersManager) getPostingsMapFBO(clientID, token string) map[string]int {
	postingsWithCountFBO := make(map[string]int)
	since := time.Now().AddDate(0, 0, m.DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	postingsListFbo := ozon.PostingsListFbo(clientID, token, since, to, 0)
	for _, posting := range postingsListFbo.Result {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBO[product.OfferID] += product.Quantity
			}
		}
	}
	return postingsWithCountFBO
}
func (m OrdersManager) GetReturnsMap(clientID, token, since, to string) (map[string]int, error) {
	var lastID int
	hasNext := true
	returnsWithCount := make(map[string]int)
	/*
		Лимит у запроса 1000, но нам нужны все возвраты,
		поэтому делаем цикл с lastID и добавляем в срез returnsFBO
	*/
	for hasNext {
		returns, err := ozon.ReturnsList(clientID, token, lastID, since, to)
		if err != nil {
			return returnsWithCount, err
		}

		for _, value := range returns.Returns {
			if value.Visual.Status.SysName == "ReturnedToOzon" {
				returnsWithCount[value.Product.OfferID] += value.Product.Quantity
			}
			lastID = value.ID
		}
		hasNext = returns.HasNext
	}

	return returnsWithCount, nil
}
