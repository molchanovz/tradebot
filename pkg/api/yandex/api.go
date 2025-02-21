package yandex

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

var campaignId = 90788543

func ShipmentInfo(token, supplyId string) (string, error) {

	url := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/first-mile/shipments/%v", campaignId, supplyId)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", token)

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

func OrderInfo(token string, orderId int64) (string, error) {

	url := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/orders/%v", campaignId, orderId)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", token)

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

func GetStickers(token string, orderId int64) (string, error) {
	url := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/orders/%v/delivery/labels?format=A9_HORIZONTALLY", campaignId, orderId)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", token)

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

func getOrdersFbo(yandexKey string, daysAgo int) (string, error) {

	date := time.Now().AddDate(0, 0, -daysAgo)

	println(date.Format("2006-01-02"))

	url := "https://api.partner.market.yandex.ru/campaigns/49152956/stats/orders"

	body := []byte(fmt.Sprintf(`{
  "dateFrom": "%v",
  "dateTo": "%v",
  "statuses": [
    "DELIVERY", "DELIVERED", "PARTIALLY_DELIVERED", "PARTIALLY_RETURNED", "PENDING", "PICKUP", "PROCESSING", "RESERVED", "UNKNOWN", "UNPAID", "LOST"
  ],
  "hasCis": false,
    "fake":false
}`, date.Format("2006-01-02"), date.Format("2006-01-02")))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", yandexKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка: получен статус %v", resp.StatusCode)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString), nil
}
