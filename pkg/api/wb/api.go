package wb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func StocksFbo(apiKey string) string {

	url := "https://statistics-api.wildberries.ru/api/v1/supplier/stocks"
	body := []byte(``)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Authorization", apiKey)
	q := req.URL.Query()
	q.Add("dateFrom", "2019-06-20")
	req.URL.RawQuery = q.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка получения ФБС заказов: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

func GetOrdersBySupplyId(wildberriesKey, supplyId string) (string, error) {

	url := "https://marketplace-api.wildberries.ru/api/v3/supplies/" + supplyId + "/orders"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", wildberriesKey) // Замените на ваш API токен, если нужен

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка: получен статус %v", resp.StatusCode)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString), nil
}

func GetCodesByOrderId(wildberriesKey string, orderId int) string {

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

	// Преобразование данных в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Ошибка при преобразовании данных в JSON:", err)
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", wildberriesKey) // Замените на ваш API токен, если нужен

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка: получен статус %d", resp.StatusCode)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

/*
API метод для получения фбс заказов
*/
func OrdersFBS(apiKey string, daysAgo int) string {

	url := "https://marketplace-api.wildberries.ru/api/v3/orders"
	body := []byte(``)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Authorization", apiKey)
	q := req.URL.Query()
	q.Add("limit", "1000")
	q.Add("next", "0")
	q.Add("dateFrom", strconv.Itoa(int(getUnix(time.Now().AddDate(0, 0, -(daysAgo+1))))))
	q.Add("dateTo", strconv.Itoa(int(getUnix(time.Now().AddDate(0, 0, -daysAgo)))))
	req.URL.RawQuery = q.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка получения ФБС заказов: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

func OrdersFBS_status(wildberriesKey string, orderId int) string {

	baseURL := "https://marketplace-api.wildberries.ru/api/v3/orders/status"

	type RequestBody struct {
		Orders []int `json:"orders"`
	}

	data := RequestBody{
		Orders: []int{orderId},
	}

	// Преобразование данных в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Ошибка при преобразовании данных в JSON:", err)
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", wildberriesKey) // Замените на ваш API токен, если нужен

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка: получен статус %d", resp.StatusCode)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

/*
API метод для получения всех заказов за дату, равную (now() - daysAgo). Максимум 1 запрос в минуту
*/
func ApiOrdersALL(apiKey string, daysAgo int) string {

	date := time.Now().AddDate(0, 0, -daysAgo)

	url := "https://statistics-api.wildberries.ru/api/v1/supplier/orders"
	body := []byte(`{}`)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Authorization", apiKey)

	q := req.URL.Query()
	q.Add("dateFrom", date.Format("2006-01-02"))
	q.Add("flag", "1")
	req.URL.RawQuery = q.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка получения ALL заказов: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

func ApiSalesAndReturns(apiKey string, daysAgo int) string {

	date := time.Now().AddDate(0, 0, -daysAgo)

	url := "https://statistics-api.wildberries.ru/api/v1/supplier/sales"
	body := []byte(`{}`)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Authorization", apiKey)

	q := req.URL.Query()
	q.Add("dateFrom", date.Format("2006-01-02"))
	q.Add("flag", "1")
	req.URL.RawQuery = q.Encode()

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка получения ALL заказов: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

func getUnix(date time.Time) int64 {
	nowStr := fmt.Sprint(date.Format("2006-01-02"), "T21:00:00")
	t, _ := time.Parse("2006-01-02T15:04:05", nowStr)
	return t.Unix()
}
