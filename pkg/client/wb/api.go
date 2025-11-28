package wb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func request(reqType, url string, headers map[string]string, params map[string]string, body []byte) (string, error) {
	req, err := http.NewRequest(reqType, url, bytes.NewBuffer(body))

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	for s := range headers {
		req.Header.Set(s, headers[s])
	}

	q := req.URL.Query()
	for s := range params {
		q.Add(s, params[s])
	}

	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get status: %v, %v", resp.Status, string(response))
	}

	return string(response), nil
}

func GET(url string, headers map[string]string, params map[string]string, body []byte) (string, error) {
	response, err := request(http.MethodGet, url, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func POST(url string, headers map[string]string, params map[string]string, body []byte) (string, error) {
	response, err := request(http.MethodPost, url, headers, params, body)
	if err != nil {
		return "", err
	}
	return response, nil
}

func stocksFbo(token string) (string, error) {
	baseURL := "https://statistics-api.wildberries.ru/api/v1/supplier/stocks"

	body := []byte(``)

	headers := map[string]string{
		"Authorization": token,
	}

	params := map[string]string{
		"dateFrom": "2019-06-20",
	}

	response, err := GET(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func getOrdersBySupplyID(token, supplyID string) (string, error) {
	baseURL := "https://marketplace-api.wildberries.ru/api/v3/supplies/" + supplyID + "/orders"
	body := []byte(``)

	headers := map[string]string{
		"Authorization": token,
	}

	response, err := GET(baseURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func getReturns(token, dateFrom, dateTo string) (string, error) {

	params := url.Values{}
	params.Add("dateFrom", dateFrom)
	params.Add("dateTo", dateTo)

	// Основной URL
	baseURL := "https://seller-analytics-api.wildberries.ru/api/v1/analytics/goods-return"

	// Добавление параметров к URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	body := []byte(``)

	headers := map[string]string{
		"Authorization": token,
	}

	response, err := GET(fullURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func getCodesByOrderID(token string, orderID int) (string, error) {
	params := url.Values{}
	params.Add("type", "png")
	params.Add("width", "58")
	params.Add("height", "40")

	// Основной URL
	baseURL := "https://marketplace-api.wildberries.ru/api/v3/orders/stickers"

	// Добавление параметров к URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	type RequestBody struct {
		Orders []int `json:"orders"`
	}

	data := RequestBody{
		Orders: []int{orderID},
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	response, err := POST(fullURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}

// Получения фбс заказов
func ordersFBS(token string, daysAgo int) (string, error) {
	baseURL := "https://marketplace-api.wildberries.ru/api/v3/orders"
	body := []byte(``)

	headers := map[string]string{
		"Authorization": token,
	}

	params := map[string]string{
		"limit":    "1000",
		"next":     "0",
		"dateFrom": strconv.Itoa(int(getUnix(time.Now().AddDate(0, 0, -(daysAgo + 1))))),
		"dateTo":   strconv.Itoa(int(getUnix(time.Now().AddDate(0, 0, -daysAgo)))),
	}

	response, err := GET(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func ordersFBSStatus(token string, orderID int) (string, error) {
	baseURL := "https://marketplace-api.wildberries.ru/api/v3/orders/status"

	type RequestBody struct {
		Orders []int `json:"orders"`
	}

	data := RequestBody{
		Orders: []int{orderID},
	}

	// Преобразование данных в JSON
	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("ошибка при преобразовании данных в JSON: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": token,
	}

	response, err := POST(baseURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}
func getCards(token string, nmID *int, updatedAt *time.Time, limit *int) (string, error) {
	baseURL := "https://content-api.wildberries.ru/content/v2/get/cards/list"

	type RequestBody struct {
		Settings struct {
			Sort struct {
				Ascending bool `json:"ascending"`
			} `json:"sort"`
			Cursor struct {
				UpdatedAt *time.Time `json:"updatedAt"`
				NmId      *int       `json:"nmID"`
				Limit     *int       `json:"limit"`
			} `json:"cursor"`
			Filter struct {
				WithPhoto int `json:"withPhoto"`
			} `json:"filter"`
		} `json:"settings"`
	}

	var data RequestBody
	data.Settings.Cursor.Limit = limit
	data.Settings.Cursor.NmId = nmID
	data.Settings.Cursor.UpdatedAt = updatedAt

	data.Settings.Sort.Ascending = true
	data.Settings.Filter.WithPhoto = -1

	if limit != nil {
		data.Settings.Cursor.Limit = limit
	}

	// Преобразование данных в JSON
	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("ошибка при преобразовании данных в JSON: %w", err)
	}

	headers := map[string]string{
		"Authorization": token,
	}

	response, err := POST(baseURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}

/*
API метод для получения всех заказов за дату, равную (now() - daysAgo). Максимум 1 запрос в минуту
*/
func apiOrdersALL(token string, daysAgo, flag int) (string, error) {
	date := time.Now().AddDate(0, 0, -daysAgo)

	baseURL := "https://statistics-api.wildberries.ru/api/v1/supplier/orders"

	body := []byte(``)

	headers := map[string]string{
		"Authorization": token,
	}

	params := map[string]string{
		"dateFrom": date.Format("2006-01-02"),
		"flag":     strconv.Itoa(flag),
	}

	response, err := GET(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func apiSalesAndReturns(token string, daysAgo int) (string, error) {
	date := time.Now().AddDate(0, 0, -daysAgo)

	baseURL := "https://statistics-api.wildberries.ru/api/v1/supplier/sales"

	body := []byte(``)

	headers := map[string]string{
		"Authorization": token,
	}

	params := map[string]string{
		"dateFrom": date.Format("2006-01-02"),
		"flag":     "1",
	}

	response, err := GET(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func getUnix(date time.Time) int64 {
	nowStr := fmt.Sprint(date.Format("2006-01-02"), "T21:00:00")
	t, _ := time.Parse("2006-01-02T15:04:05", nowStr)
	return t.Unix()
}
