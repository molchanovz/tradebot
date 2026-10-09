package yandex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

// ordersStatsLimit — максимум заказов на странице stats/orders (по умолчанию API отдаёт 100).
const ordersStatsLimit = 200

// getOrders возвращает страницу заказов кампании, оформленных в день date.
// pageToken — токен страницы из предыдущего ответа, для первой страницы пустой.
func getOrders(ctx context.Context, campaignID, yandexKey string, date time.Time, pageToken string) (string, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(ordersStatsLimit))
	if pageToken != "" {
		params.Set("page_token", pageToken)
	}
	u := fmt.Sprintf("https://api.partner.market.yandex.ru/campaigns/%v/stats/orders?%s", campaignID, params.Encode())

	body := []byte(fmt.Sprintf(`{
  "dateFrom": "%v",
  "dateTo": "%v",
  "statuses": [
    "DELIVERY", "DELIVERED", "PARTIALLY_DELIVERED", "PARTIALLY_RETURNED", "PENDING", "PICKUP", "PROCESSING", "RESERVED", "UNKNOWN", "UNPAID", "LOST"
  ],
  "hasCis": false,
    "fake":false
}`, date.Format("2006-01-02"), date.Format("2006-01-02")))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", yandexKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	jsonString, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка: получен статус %v: %s", resp.StatusCode, jsonString)
	}

	return string(jsonString), nil
}
