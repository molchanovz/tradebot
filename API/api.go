package API

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

func stocksFbo(apiKey string) string {

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
