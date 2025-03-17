package ozon_orders_returns

import (
	"WildberriesGo_bot/pkg/api/ozon"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	DaysAgo       = 1
	spreadsheetId = "13VI6x59ht10mMlfH3F8n5_jXbXB1kypBDH0vIDnsb3M"
)

// WriteToGoogleSheets Заполнение гугл таблицы с id = spreadsheetId
func WriteToGoogleSheets(ClientId string, ApiKey string) {
	date := time.Now().AddDate(0, 0, -DaysAgo)
	sheetsName := "Заказы OZON-" + strconv.Itoa(date.Day())

	var values [][]interface{}
	values = append(values, []interface{}{"Отчет за " + date.Format("02.01.2006")})

	writeRange := sheetsName + "!A1"
	write(spreadsheetId, writeRange, values)

	//Заполнение заказов FBS в writeRange
	postingsWithCountFBS := getPostingsMapFBS(ClientId, ApiKey)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBS"})
	for article, count := range postingsWithCountFBS {
		values = append(values, []interface{}{article, count})
	}
	writeRange = sheetsName + "!A2:B100"
	write(spreadsheetId, writeRange, values)

	//Заполнение заказов FBO в writeRange
	postingsWithCountFBO := getPostingsMapFBO(ClientId, ApiKey)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Заказы FBO"})
	for article, count := range postingsWithCountFBO {
		values = append(values, []interface{}{article, count})
	}
	writeRange = sheetsName + "!D2:E100"
	write(spreadsheetId, writeRange, values)

	since := time.Now().AddDate(0, 0, DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	//Заполнение возвратов
	println(since)
	println(to)
	returnsWithCount := getReturnsMap(ClientId, ApiKey, since, to)
	values = [][]interface{}{}
	values = append(values, []interface{}{"Возвраты"})
	for article, count := range returnsWithCount {
		values = append(values, []interface{}{article, count})
	}
	writeRange = sheetsName + "!G2:H100"
	write(spreadsheetId, writeRange, values)
}

func getPostingsMapFBS(ClientId string, ApiKey string) map[string]int {
	postingsWithCountFBS := make(map[string]int)
	since := time.Now().AddDate(0, 0, DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	potingsListFbs := ozon.PostingsListFbs(ClientId, ApiKey, since, to, 0)
	for _, posting := range potingsListFbs.Result.PostingsFBS {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBS[product.OfferId] += product.Quantity
			}
		}
	}
	return postingsWithCountFBS
}
func getPostingsMapFBO(ClientId string, ApiKey string) map[string]int {
	postingsWithCountFBO := make(map[string]int)
	since := time.Now().AddDate(0, 0, DaysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, DaysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	postings_list_fbo := ozon.PostingsListFbo(ClientId, ApiKey, since, to, 0)
	for _, posting := range postings_list_fbo.Result {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBO[product.OfferId] += product.Quantity
			}
		}
	}
	return postingsWithCountFBO
}

func getReturnsMap(ClientId, ApiKey, since, to string) map[string]int {
	LastID := 0
	hasNext := true
	returnsWithCount := make(map[string]int)
	var returns ozon.Returns
	/*
		Лимит у запроса 1000, но нам нужны все возвраты,
		поэтому делаем цикл с LastID и добавляем в срез returnsFBO
	*/
	for hasNext {
		returns = ozon.ReturnsList(ClientId, ApiKey, LastID, since, to)
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

func initEnv(path, name string) (string, error) {
	err := godotenv.Load(path)
	if err != nil {
		log.Printf("Ошибка загрузки файла %s: %v\n", path, err)
		return "", fmt.Errorf("ошибка загрузки файла " + path)
	}
	// Получаем значения переменных среды
	env := os.Getenv(name)

	if env == "" {
		return "", fmt.Errorf("переменная среды " + name + " не установлена")
	}
	return env, err
}
