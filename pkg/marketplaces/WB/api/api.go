package api

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
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
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
		return "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("получен статус %v", resp.Status)
	}

	jsonString, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func GET(url string, headers map[string]string, params map[string]string, body []byte) (string, error) {
	response, err := request("GET", url, headers, params, body)
	if err != nil {
		return "", err
	}
	return response, nil
}

func POST(url string, headers map[string]string, params map[string]string, body []byte) (string, error) {
	response, err := request("POST", url, headers, params, body)
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

func getOrdersBySupplyId(token, supplyId string) (string, error) {

	baseURL := "https://marketplace-api.wildberries.ru/api/v3/supplies/" + supplyId + "/orders"
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

func getCodesByOrderId(token string, orderId int) (string, error) {

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
		Orders: []int{orderId},
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

func ordersFBSStatus(token string, orderId int) (string, error) {

	baseURL := "https://marketplace-api.wildberries.ru/api/v3/orders/status"

	type RequestBody struct {
		Orders []int `json:"orders"`
	}

	data := RequestBody{
		Orders: []int{orderId},
	}

	// Преобразование данных в JSON
	body, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Ошибка при преобразовании данных в JSON:", err)
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
