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
	daysAgo       = 1
	spreadsheetId = "138vQwwc5g3aGFbZXRBnC8U1c1pecRqD3VI6LhQ8QbKI"
)

// WriteToGoogleSheets Заполнение гугл таблицы с id = spreadsheetId
func WriteToGoogleSheets(ClientId string, ApiKey string) {
	date := time.Now().AddDate(0, 0, -daysAgo)
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

	//Заполнение возвратов
	returnsWithCount := getReturnsMap(ClientId, ApiKey)
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
	since := time.Now().AddDate(0, 0, daysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, daysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	potingsListFbs := ozon.PostingsListFbs(ClientId, ApiKey, since, to)
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
	since := time.Now().AddDate(0, 0, daysAgo*(-1)-1).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, daysAgo*(-1)).Format("2006-01-02") + "T21:00:00.000Z"
	postings_list_fbo := ozon.PostingsListFbo(ClientId, ApiKey, since, to)
	for _, posting := range postings_list_fbo.Result {
		if posting.Status != "cancelled" {
			for _, product := range posting.Products {
				postingsWithCountFBO[product.OfferId] += product.Quantity
			}
		}
	}
	return postingsWithCountFBO
}

func getReturnsMap(ClientId string, ApiKey string) map[string]int {
	LastID := 0
	returnsWithCount := make(map[string]int)
	/*
		Лимит у запроса 1000, но нам нужны все возвраты,
		поэтому делаем цикл с LastID и добавляем в срез returnsFBO
	*/
	returns_fbo, LastID := ozon.ReturnsFbo(ClientId, ApiKey, LastID)
	returnsFBO := make([]ozon.ReturnFBO, 0, len(returns_fbo))
	returnsFBO = append(returnsFBO, returns_fbo...)
	for LastID != 0 {
		returns_fbo, LastID = ozon.ReturnsFbo(ClientId, ApiKey, LastID)
		returnsFBO = append(returnsFBO, returns_fbo...)
	}
	for i := range returnsFBO {
		parsedTime := dateParser(returnsFBO[i].ReturnedToOzonMoment)
		// Получаем год, месяц и день
		year := parsedTime.Year()
		month := parsedTime.Month()
		day := parsedTime.Day()

		if year == time.Now().Year() && month == time.Now().Month() && day == time.Now().Day()-daysAgo {
			posting := ozon.PostingFbo(ClientId, ApiKey, returnsFBO[i].PostingNumber)
			for _, product := range posting.Result.Products {
				returnsWithCount[product.OfferId] += product.Quantity
			}
		}
	}

	LastID = 0

	returns_fbs, LastID := ozon.ReturnsFbs(ClientId, ApiKey, LastID)
	returnsListFBS := make([]ozon.ReturnFBS, 0, len(returns_fbs))
	returnsListFBS = append(returnsListFBS, returns_fbs...)

	for LastID != 0 {
		returns_fbs, LastID = ozon.ReturnsFbs(ClientId, ApiKey, LastID)
		returnsListFBS = append(returnsListFBS, returns_fbs...)
	}

	for i := range returnsListFBS {
		parsedTime := dateParser(returnsListFBS[i].ReturnDate)
		year := parsedTime.Year()
		month := parsedTime.Month()
		day := parsedTime.Day()
		if year == time.Now().Year() && month == time.Now().Month() && day == time.Now().Day()-daysAgo {
			posting := ozon.PostingFbs(ClientId, ApiKey, returnsListFBS[i].PostingNumber)
			for _, product := range posting.Result.Products {
				returnsWithCount[product.OfferId] += product.Quantity
			}

		}
	}
	return returnsWithCount
}

func dateParser(date string) time.Time {

	parsedTime, err := time.Parse(time.RFC3339, date)
	if err != nil {
		fmt.Println("Ошибка разбора даты:", err)
		return time.Time{}
	}
	return parsedTime
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
