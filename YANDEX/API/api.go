package API

import (
	"fmt"
	"io"
	"net/http"
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
