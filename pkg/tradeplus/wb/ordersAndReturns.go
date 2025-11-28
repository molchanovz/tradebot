package wb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tradebot/pkg/client/wb"
	"tradebot/pkg/tradeplus"
)

type OrdersManager struct {
	tradeplus.OrderManager
	client wb.Client
}

func NewOrdersManager(token, spreadsheetID string) OrdersManager {
	manager := OrdersManager{tradeplus.NewOrdersManager(spreadsheetID), wb.NewClient(token)}
	return manager
}

func (m OrdersManager) Write() error {
	date := time.Now().AddDate(0, 0, -m.DaysAgo)
	sheetsName := "Заказы WB-" + strconv.Itoa(date.Day())

	var values [][]interface{}

	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	writeRange := sheetsName + "!A1"

	err := m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return err
	}

	// Запись ALL заказов
	postingsWithCountFBO, postingsWithCountFBS, err := m.getPostingsMap()
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

	// Запись FBS заказов
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

	// Запись возвратов
	returnsWithCount := m.getReturnsMap()
	writeRange = sheetsName + "!G2:H100"
	colName = "Возвраты"
	values = [][]interface{}{}
	values = append(values, []interface{}{colName})
	for article, count := range returnsWithCount {
		values = append(values, []interface{}{article, count})
	}
	err = m.GoogleService.Write(m.SpreadsheetID, writeRange, values)
	if err != nil {
		return err
	}

	return nil
}

func (m OrdersManager) getPostingsMap() (map[string]int, map[string]int, error) {
	postingsWithCountFBO := make(map[string]int)
	postingsWithCountFBS := make(map[string]int)

	postingsList, err := m.client.GetAllOrders(m.DaysAgo, 1)
	if err != nil {
		return nil, nil, fmt.Errorf("wb getAllOrders failed: %w", err)
	}

	for _, posting := range postingsList {
		if !posting.IsCancel {
			switch posting.WarehouseType {
			case "Склад WB":
				postingsWithCountFBO[posting.SupplierArticle]++
			case "Склад продавца":
				postingsWithCountFBS[posting.SupplierArticle]++
			default:
				if strings.Contains(posting.WarehouseName, "мп") {
					postingsWithCountFBS[posting.SupplierArticle]++
				} else {
					postingsWithCountFBO[posting.SupplierArticle]++
				}
			}
		}
	}

	return postingsWithCountFBO, postingsWithCountFBS, nil
}

func (m OrdersManager) getReturnsMap() map[string]int {
	returnsWithCount := make(map[string]int)

	returnsList, _ := m.client.GetSalesAndReturns(m.DaysAgo)
	for _, someReturn := range returnsList {
		if strings.HasPrefix(someReturn.SaleID, "R") {
			returnsWithCount[someReturn.SupplierArticle]++
		}
	}

	return returnsWithCount
}
