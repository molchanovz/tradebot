package OrdersAndReturns

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tradebot/pkg/api/wb"
	"tradebot/pkg/google"
)

type WbManager struct {
	daysAgo              int
	spreadsheetId, token string
	googleSheets         google.SheetsService
}

func NewWbManager(token, spreadsheetId string, daysAgo int) WbManager {
	return WbManager{
		token:         token,
		daysAgo:       daysAgo,
		spreadsheetId: spreadsheetId,
		googleSheets:  google.NewSheetsService("token.json", "credentials.json"),
	}
}

func (m WbManager) WriteToGoogleSheets() error {
	date := time.Now().AddDate(0, 0, -m.daysAgo)
	sheetsName := "Заказы WB-" + strconv.Itoa(date.Day())

	var values [][]interface{}

	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	writeRange := sheetsName + "!A1"

	err := m.googleSheets.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	//Запись ALL заказов
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
	err = m.googleSheets.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	//Запись FBS заказов
	writeRange = sheetsName + "!D2:E100"
	colName = "Заказы FBS"
	values = [][]interface{}{}
	values = append(values, []interface{}{colName})
	for article, count := range postingsWithCountFBS {
		values = append(values, []interface{}{article, count})
	}
	err = m.googleSheets.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	//Запись возвратов
	returnsWithCount := m.getReturnsMap()
	writeRange = sheetsName + "!G2:H100"
	colName = "Возвраты"
	values = [][]interface{}{}
	values = append(values, []interface{}{colName})
	for article, count := range returnsWithCount {
		values = append(values, []interface{}{article, count})
	}
	err = m.googleSheets.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	return nil
}

func (m WbManager) ordersMapFBS() map[string]int {
	postingsWithCountFBS := make(map[string]int)
	postingsList := wb.GetOrdersFBS(m.token, m.daysAgo)

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
		status := wb.GetPostingStatus(m.token, posting.Id)
		if !isOrderCanceled(status) {
			postingsWithCountFBS[posting.Article] += 1
		}
	}
	return postingsWithCountFBS
}
func (m WbManager) getPostingsMap() (map[string]int, map[string]int, error) {
	postingsWithCountFBO := make(map[string]int)
	postingsWithCountFBS := make(map[string]int)
	postingsList := wb.GetAllOrders(m.token, m.daysAgo, 1)

	for _, posting := range postingsList {
		if posting.IsCancel == false {
			if posting.WarehouseType == "Склад WB" {
				postingsWithCountFBO[posting.SupplierArticle]++
			} else if posting.WarehouseType == "Склад продавца" {
				postingsWithCountFBS[posting.SupplierArticle]++
			} else {
				return postingsWithCountFBO, postingsWithCountFBS, fmt.Errorf("неопознанный тип склада: %v", posting.WarehouseType)
			}

		} else {
			fmt.Println("Пропавший заказ ", posting.SupplierArticle, " ", posting.IsCancel)
		}
	}
	return postingsWithCountFBO, postingsWithCountFBS, nil
}
func (m WbManager) getReturnsMap() map[string]int {
	returnsWithCount := make(map[string]int)
	returnsList := wb.GetSalesAndReturns(m.token, m.daysAgo)
	for _, someReturn := range returnsList {
		if strings.HasPrefix(someReturn.SaleID, "R") {
			returnsWithCount[someReturn.SupplierArticle]++
		}
	}
	return returnsWithCount
}
