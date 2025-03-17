package ozon

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

/*
Метод получения ФБО заказа исходя из posting_number (нужен для извлечения товаров в возврате)
*/
func V2PostingFboGet(ClientId, ApiKey, PostingNumber string) string {

	url := "https://api-seller.ozon.ru/v2/posting/fbo/get"
	body := []byte(`{
  "posting_number": "` + PostingNumber + `",
  "translit": true,
  "with": {
    "analytics_data": false,
    "financial_data": false
  }
}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка v2_posting_fbo_get: получен статус %s", resp.Status)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}

// V3ReturnsCompanyFbo метод получения ФБО возвратов со статусом ReturnedToOzon. Получаем возвраты только тогда, когда возврат приедет на склад озон
func V3ReturnsCompanyFbo(ClientId, ApiKey string, daysAgo, LastID int) string {

	url := "https://api-seller.ozon.ru/v3/returns/company/fbo"
	body := []byte(`{
  "filter": {
    "status": [
      "ReturnedToOzon"
    ]
  },
  "last_id":` + strconv.Itoa(LastID) + `,
  "limit": 1000
}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка v3_returns_company_fbo: получен статус %s", resp.Status)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}
func returnsList(ClientId, ApiKey string, LastID int, since, to string) string {

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
}`, since, to, LastID))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка ReturnsList: получен статус %s", resp.Status)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}

// V3ReturnsCompanyFbs метод получения ФБС возвратов со статусом moving_to_resale
func V3ReturnsCompanyFbs(ClientId, ApiKey string, LastID int) string {

	url := "https://api-seller.ozon.ru/v3/returns/company/fbs"
	body := []byte(`{
  "filter": {
    "status": "moving_to_resale"
  },
  "limit": 1000,
  "last_id": ` + strconv.Itoa(LastID) + `
}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка v3_returns_company_fbs: получен статус %s", resp.Status)
	}

	jsonString, _ := io.ReadAll(resp.Body)

	return string(jsonString)
}

// V3PostingFbsGet метод получения ФБС заказа исходя из posting_number (нужен для извлечения товаров в возврате)
func V3PostingFbsGet(ClientId, ApiKey, PostingNumber string) string {

	url := "https://api-seller.ozon.ru/v3/posting/fbs/get"
	body := []byte(`{
  	"posting_number": "` + PostingNumber + `",
  	"with": {
    	"analytics_data": false,
    	"barcodes": false,
    	"financial_data": false,
    	"product_exemplars": false,
    	"translit": false
 	 }
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка v3_returns_company_fbs: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

// V3PostingFbsList метод получения ФБС заказов
func V3PostingFbsList(ClientId, ApiKey, since, to string, offset int) string {

	url := "https://api-seller.ozon.ru/v3/posting/fbs/list"
	body := []byte(fmt.Sprintf(`{
  "dir": "ASC",
  "filter": {
    "since": "%v",
    "to": "%v"
},
  "limit": 1000,
  "offset": %v,
  "with": {
    "analytics_data": false,
    "barcodes": false,
    "financial_data": true,
    "translit": false
  }
}`, since, to, offset))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка v2_posting_fbo_get: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

// V2PostingFboList метод получения ФБО заказов
func V2PostingFboList(ClientId, ApiKey, since, to string, offset int) string {

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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа

	if resp.StatusCode != http.StatusOK {
		errString, _ := io.ReadAll(resp.Body)
		log.Fatalf("Ошибка v2_posting_fbo_get: получен статус %s. %s", resp.Status, errString)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}
func v2StockOnWarehouses(ClientId, ApiKey string) string {

	url := "https://api-seller.ozon.ru/v2/analytics/stock_on_warehouses"
	body := []byte(`{
  		"limit": 1000,
  		"offset": 0,
  		"warehouse_type": "ALL"
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка v2_posting_fbo_get: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}
func v1Clusters(ClientId, ApiKey string) string {

	url := "https://api-seller.ozon.ru/v1/cluster/list"
	body := []byte(`{
  		"cluster_type": "CLUSTER_TYPE_OZON"
	}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	// Устанавливаем необходимые заголовки (если нужны)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", ClientId)
	req.Header.Set("Api-Key", ApiKey)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка v2_posting_fbo_get: получен статус %s", resp.Status)
	}

	// Читаем тело ответа
	jsonString, _ := io.ReadAll(resp.Body)

	// Выводим ответ
	return string(jsonString)
}

//func v1ReportPostingsCreate(Client_Id, Api_Key string) string {
//
//	url := "https://api-seller.ozon.ru/v1/report/postings/create"
//	body := []byte(`{
//  "filter": {
//    "processed_at_from": "` + time.Now().AddDate(0, 0, -1).Format("2006-01-02T15:04:05.000Z") + `",
//    "processed_at_to": "` + time.Now().Format("2006-01-02T15:04:05.000Z") + `",
//    "delivery_schema": [
//      "fbo"
//    ],
//    "sku": [],
//    "cancel_reason_id": [],
//    "offer_id": "",
//    "status_alias": [],
//    "statuses": [],
//    "title": ""
//  },
//  "language": "DEFAULT"
//}`)
//
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
//
//	if err != nil {
//		log.Fatalf("Ошибка создания запроса: %v", err)
//	}
//
//	// Устанавливаем необходимые заголовки (если нужны)
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Client-Id", Client_Id)
//	req.Header.Set("Api-Key", Api_Key)
//
//	// Выполняем запрос
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		log.Fatalf("Ошибка выполнения запроса: %v", err)
//	}
//	defer resp.Body.Close()
//
//	// Проверяем статус ответа
//	if resp.StatusCode != http.StatusOK {
//		log.Fatalf("Ошибка v2_posting_fbo_get: получен статус %s", resp.Status)
//	}
//
//	// Читаем тело ответа
//	jsonString, _ := io.ReadAll(resp.Body)
//
//	// Выводим ответ
//	return string(jsonString)
//}
//func v1ReportInfo(Client_Id, Api_Key string) string {
//	jsonString := v1ReportPostingsCreate(Client_Id, Api_Key)
//	var response ReportResponse
//
//	// Парсим JSON-ответ
//	err := json.Unmarshal([]byte(jsonString), &response)
//	if err != nil {
//		log.Fatalf("Ошибка при парсинге JSON: %v", err)
//	}
//
//	// Получаем значение поля code
//	code := response.Result.Code
//
//	url := "https://api-seller.ozon.ru/v1/report/info"
//	body := []byte(`{
//  	"code": "` + code + `"
//	}`)
//
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
//
//	if err != nil {
//		log.Fatalf("Ошибка создания запроса: %v", err)
//	}
//
//	// Устанавливаем необходимые заголовки (если нужны)
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Client-Id", Client_Id)
//	req.Header.Set("Api-Key", Api_Key)
//
//	// Выполняем запрос
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		log.Fatalf("Ошибка выполнения запроса: %v", err)
//	}
//	defer resp.Body.Close()
//
//	// Проверяем статус ответа
//	if resp.StatusCode != http.StatusOK {
//		log.Fatalf("Ошибка v2_posting_fbo_get: получен статус %s", resp.Status)
//	}
//
//	// Читаем тело ответа
//	jsonBytes, _ := io.ReadAll(resp.Body)
//	jsonString = string(jsonBytes)
//	// Выводим ответ
//	return string(jsonString)
//}
