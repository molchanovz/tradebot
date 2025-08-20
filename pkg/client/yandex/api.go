package yandex

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

var campaignID = 90788543

func ShipmentInfo(token, supplyID string) (string, error) {
	url := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/first-mile/shipments/%v", campaignID, supplyID)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", token)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса ShipmentInfo: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка ShipmentInfo: получен статус %v", resp.StatusCode)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString), nil
}

func OrderInfo(token string, orderID int64) (string, error) {
	url := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/orders/%v", campaignID, orderID)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", token)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
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

func GetStickers(token string, orderID int64) (string, error) {
	url := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/orders/%v/delivery/labels?format=A9_HORIZONTALLY", campaignID, orderID)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", token)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
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

func getOrders(campaignID, yandexKey string, daysAgo int) (string, error) {
	date := time.Now().AddDate(0, 0, -daysAgo)

	url := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/stats/orders", campaignID)

	body := []byte(fmt.Sprintf(`{
  "dateFrom": "%v",
  "dateTo": "%v",
  "statuses": [
    "DELIVERY", "DELIVERED", "PARTIALLY_DELIVERED", "PARTIALLY_RETURNED", "PENDING", "PICKUP", "PROCESSING", "RESERVED", "UNKNOWN", "UNPAID", "LOST"
  ],
  "hasCis": false,
    "fake":false
}`, date.Format("2006-01-02"), date.Format("2006-01-02")))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", yandexKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка: получен статус %v", resp.StatusCode)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString), nil
}
