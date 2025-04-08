package OrdersAndReturns

import (
	"WildberriesGo_bot/pkg/api/ozon"
	"WildberriesGo_bot/pkg/googleService"
	"strconv"
	"time"
)

type OzonManager struct {
	daysAgo                        int
	spreadsheetId, clientId, token string
	googleService                  googleService.GoogleService
}

func NewOzonManager(clientId, token, spreadsheetId string, daysAgo int) OzonManager {
	return OzonManager{
		clientId:      clientId,
		token:         token,
		daysAgo:       daysAgo,
		spreadsheetId: spreadsheetId,
		googleService: googleService.NewGoogleService("token.json", "credentials.json"),
	}
}

// WriteToGoogleSheets Заполнение гугл таблицы с id = spreadsheetId
func (m OzonManager) WriteToGoogleSheets() error {
	date := time.Now().AddDate(0, 0, -m.daysAgo)
	sheetsName := "Заказы OZON-" + strconv.Itoa(date.Day())

	var values [][]interface{}
	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	writeRange := sheetsName + "!A1"

	err := m.googleService.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	//Заполнение заказов FBS в writeRange
	postingsWithCountFBS := m.getPostingsMapFBS()
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBS"})
	for article, count := range postingsWithCountFBS {
		values = append(values, []interface{}{article, count})
	}
	writeRange = sheetsName + "!A2:B100"
	err = m.googleService.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	postingsWithCountFBO := m.getPostingsMapFBO()
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBO"})
	for article, count := range postingsWithCountFBO {
		values = append(values, []interface{}{article, count})
	}
	writeRange = sheetsName + "!D2:E100"
	err = m.googleService.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	since := time.Now().AddDate(0, 0, m.daysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.daysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	//Заполнение возвратов
	returnsWithCount := m.getReturnsMap(since, to)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Возвраты"})
	for article, count := range returnsWithCount {
		values = append(values, []interface{}{article, count})
	}
	writeRange = sheetsName + "!G2:H100"
	err = m.googleService.Write(m.spreadsheetId, writeRange, values)
	if err != nil {
		return err
	}

	return nil
}

func (m OzonManager) getPostingsMapFBS() map[string]int {
	postingsWithCountFBS := make(map[string]int)
	since := time.Now().AddDate(0, 0, m.daysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.daysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	potingsListFbs := ozon.PostingsListFbs(m.clientId, m.token, since, to, 0, "")
	for _, posting := range potingsListFbs.Result.PostingsFBS {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBS[product.OfferId] += product.Quantity
			}
		}
	}
	return postingsWithCountFBS
}
func (m OzonManager) getPostingsMapFBO() map[string]int {
	postingsWithCountFBO := make(map[string]int)
	since := time.Now().AddDate(0, 0, m.daysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, m.daysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	postings_list_fbo := ozon.PostingsListFbo(m.clientId, m.token, since, to, 0)
	for _, posting := range postings_list_fbo.Result {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBO[product.OfferId] += product.Quantity
			}
		}
	}
	return postingsWithCountFBO
}
func (m OzonManager) getReturnsMap(since, to string) map[string]int {
	LastID := 0
	hasNext := true
	returnsWithCount := make(map[string]int)
	var returns ozon.Returns
	/*
		Лимит у запроса 1000, но нам нужны все возвраты,
		поэтому делаем цикл с LastID и добавляем в срез returnsFBO
	*/
	for hasNext {
		returns = ozon.ReturnsList(m.clientId, m.token, LastID, since, to)
		for _, value := range returns.Returns {
			if value.Visual.Status.SysName == "ReturnedToOzon" {
				returnsWithCount[value.Product.OfferId] += value.Product.Quantity
			}
			LastID = value.Id
		}
		hasNext = returns.HasNext
	}
	return returnsWithCount
}
