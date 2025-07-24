package ozon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type OzonClient struct {
}

// v2PostingFboGet для получения ФБО заказа исходя из posting_number (нужен для извлечения товаров в возврате)
func v2PostingFboGet(clientID, apiKey, postingNumber string) (string, error) {
	url := "https://api-seller.ozon.ru/v2/posting/fbo/get"
	body := []byte(`{
  "posting_number": "` + postingNumber + `",
  "translit": true,
  "with": {
    "analytics_data": false,
    "financial_data": false
  }
}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("ошибка создания запроса: %v", err)
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка v2PostingFboGet: получен статус %s", resp.Status)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString), nil
}

// V3ReturnsCompanyFbo метод получения ФБО возвратов со статусом ReturnedToOzon. Получаем возвраты только тогда, когда возврат приедет на склад озон
func V3ReturnsCompanyFbo(clientID, apiKey string, daysAgo, lastID int) string {
	url := "https://api-seller.ozon.ru/v3/returns/company/fbo"
	body := []byte(`{
  "filter": {
    "status": [
      "ReturnedToOzon"
    ]
  },
  "last_id":` + strconv.Itoa(lastID) + `,
  "limit": 1000
}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v3_returns_company_fbo: получен статус %s", resp.Status)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}
func returnsList(clientID, apiKey string, lastID int, since, to string) string {
	url := "https://api-seller.ozon.ru/v1/returns/list"
	body := []byte(fmt.Sprintf(`{
  "filter": {
    "visual_status_change_moment": {
      "time_from": "%v",
      "time_to": "%v"
    }
  },
  "limit": 500,
  "last_id": %v
}`, since, to, lastID))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка ReturnsList: получен статус %s", resp.Status)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}

// V3ReturnsCompanyFbs метод получения ФБС возвратов со статусом moving_to_resale
func V3ReturnsCompanyFbs(clientID, apiKey string, lastID int) string {
	url := "https://api-seller.ozon.ru/v3/returns/company/fbs"
	body := []byte(`{
  "filter": {
    "status": "moving_to_resale"
  },
  "limit": 1000,
  "last_id": ` + strconv.Itoa(lastID) + `
}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v3_returns_company_fbs: получен статус %s", resp.Status)
		return ""
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}

// V3PostingFbsGet метод получения ФБС заказа исходя из posting_number (нужен для извлечения товаров в возврате)
func V3PostingFbsGet(clientID, apiKey, postingNumber string) string {
	url := "https://api-seller.ozon.ru/v3/posting/fbs/get"
	body := []byte(`{
  	"posting_number": "` + postingNumber + `",
  	"with": {
    	"analytics_data": false,
    	"barcodes": false,
    	"financial_data": false,
    	"product_exemplars": false,
    	"translit": false
 	 }
	}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v3_returns_company_fbs: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

// V3PostingFbsList метод получения ФБС заказов
func V3PostingFbsList(clientID, apiKey, since, to string, offset int, status string) (string, error) {
	url := "https://api-seller.ozon.ru/v3/posting/fbs/list"
	body := []byte(fmt.Sprintf(`{
  "dir": "ASC",
  "filter": {
    "since": "%v",
    "to": "%v",
	"status": "%v"
},
  "limit": 1000,
  "offset": %v,
  "with": {
    "analytics_data": false,
    "barcodes": false,
    "financial_data": true,
    "translit": false
  }
}`, since, to, status, offset))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return "", err
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка V3PostingFbsList: получен статус %v", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString), nil
}
func V2PostingFbsPackageLabel(clientID, apiKey, postingNumber string) string {
	url := "https://api-seller.ozon.ru/v2/posting/fbs/package-label"
	body := []byte(fmt.Sprintf(`{
  "posting_number": [
    "%v"
  ]
}`, postingNumber))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка V2PostingFbsPackageLabel: получен статус %s. %v", resp.Status, string(jsonString))
	}

	// Выводим ответ
	return string(jsonString)
}

// V2PostingFboList метод получения ФБО заказов
func V2PostingFboList(clientID, apiKey, since, to string, offset int) string {
	url := "https://api-seller.ozon.ru/v2/posting/fbo/list"
	body := []byte(fmt.Sprintf(`{
  "dir": "ASC",
  "filter": {
    "since": "%v",
    "to": "%v"
  },
  "limit": 1000,
  "offset": %v,
  "translit": false,
  "with": {
    "analytics_data": false,
    "financial_data": true
  }
}`, since, to, offset))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errString, _ := io.ReadAll(resp.Body)
		log.Printf("Ошибка V2PostingFboList: получен статус %s. %s", resp.Status, errString)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}
func v2StockOnWarehouses(clientID, apiKey string) string {
	url := "https://api-seller.ozon.ru/v2/analytics/stock_on_warehouses"
	body := []byte(`{
  		"limit": 1000,
  		"offset": 0,
  		"warehouse_type": "ALL"
	}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v2StockOnWarehouses: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}
func v1Clusters(clientID, apiKey string) string {
	url := "https://api-seller.ozon.ru/v1/cluster/list"
	body := []byte(`{
  		"cluster_type": "CLUSTER_TYPE_OZON"
	}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v1Clusters: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

func v1AnalyticsStocks(clientID, apiKey string, skus []string) string {
	skusJSON, _ := json.Marshal(skus)

	url := "https://api-seller.ozon.ru/v1/analytics/stocks"
	body := []byte(fmt.Sprintf(`{
  "skus": %v
}`, string(skusJSON)))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v1AnalyticsStocks: получен статус %s: %v", resp.Status, string(jsonString))
	}

	// Выводим ответ
	return string(jsonString)
}

func v3ProductList(clientID, apiKey string) string {
	url := "https://api-seller.ozon.ru/v3/product/list"
	body := []byte(`{
  "filter": {
    "visibility": "TO_SUPPLY"
  },
  "last_id": "",
  "limit": 1000
}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v3ProductList: получен статус %s: %v", resp.Status, string(jsonString))
	}

	// Выводим ответ
	return string(jsonString)
}

func v4ProductInfoAttributes(clientID, apiKey string) string {
	url := "https://api-seller.ozon.ru/v4/product/info/attributes"
	body := []byte(`{
  "filter": {
    "visibility": "VISIBLE"
  },
  "limit": 1000,
  "sort_dir": "ASC"
}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return ""
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Api-Key", apiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка v4ProductInfoAttributes: получен статус %s: %v", resp.Status, string(jsonString))
	}

	// Выводим ответ
	return string(jsonString)
}
