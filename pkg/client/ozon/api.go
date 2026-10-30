package ozon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"tradebot/pkg/client/apierr"
)

// apiURL — адрес Seller API; в тестах клиента подменяется адресом тестового сервера.
const apiURL = "https://api-seller.ozon.ru"

type Client struct {
	hc               *http.Client
	baseURL          string
	clientID, apiKey string
	onRateLimit      func(retry int, delay time.Duration)
}

func NewClient(clientID, apiKey string) Client {
	return Client{
		hc: &http.Client{
			Timeout: time.Second * 10,
		},
		baseURL:  apiURL,
		clientID: clientID,
		apiKey:   apiKey,
	}
}

// WithBaseURL возвращает копию клиента, которая ходит на baseURL вместо Seller API.
// Нужен тестам, которые поднимают свой сервер вместо Озона.
func (c Client) WithBaseURL(baseURL string) Client {
	c.baseURL = baseURL
	return c
}

// WithRateLimitHook возвращает копию клиента, которая перед каждым повтором после 429
// вызывает f с номером повтора (с единицы) и паузой до него.
func (c Client) WithRateLimitHook(f func(retry int, delay time.Duration)) Client {
	c.onRateLimit = f
	return c
}

const (
	// RateLimitRetries — сколько раз повторяем запрос после 429: Озон ограничивает
	// число запросов в секунду на кабинет, и через секунду-две запрос обычно проходит.
	RateLimitRetries = 3
	// rateLimitJitter — случайная добавка к паузе, чтобы параллельные запросы
	// не повторялись в одну и ту же секунду.
	rateLimitJitter = 500 * time.Millisecond
)

// sleep подменяется в тестах, чтобы не ждать паузы между повторами.
var sleep = time.Sleep

// RateLimitDelay — пауза перед повтором номер attempt (с нуля) после 429:
// сколько просит Озон в Retry-After, иначе 1, 2, 3 секунды.
func RateLimitDelay(attempt int, h http.Header) time.Duration {
	jitter := rand.N(rateLimitJitter)
	if s, err := strconv.Atoi(h.Get("Retry-After")); err == nil && s > 0 {
		return time.Duration(s)*time.Second + jitter
	}
	return time.Duration(attempt+1)*time.Second + jitter
}

// RateLimitHeaders — заголовки ответа, похожие на сведения о лимите (Retry-After,
// X-RateLimit-* и т.п.), одной строкой для лога. Озон в документации их не описывает,
// так что по логу 429 видно, присылает ли он что-нибудь. Если таких нет, перечисляет
// имена всех заголовков.
func RateLimitHeaders(h http.Header) string {
	var found []string
	names := make([]string, 0, len(h))
	for name, values := range h {
		names = append(names, name)
		n := strings.ToLower(name)
		if strings.Contains(n, "limit") || strings.Contains(n, "retry") || strings.Contains(n, "reset") ||
			strings.Contains(n, "quota") || strings.Contains(n, "throttl") {
			found = append(found, name+"="+strings.Join(values, ","))
		}
	}
	if len(found) == 0 {
		slices.Sort(names)
		return "нет, заголовки ответа: " + strings.Join(names, ", ")
	}
	slices.Sort(found)
	return strings.Join(found, "; ")
}

func (c Client) request(reqType, baseURL string, headers map[string]string, params map[string]string, body []byte) (string, error) {
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest(reqType, baseURL, bytes.NewBuffer(body))
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

		resp, err := c.hc.Do(req)
		if err != nil {
			return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Printf("ozon %s: 429, попытка %d, заголовки лимита: %s", req.URL.Path, attempt+1, RateLimitHeaders(resp.Header))
		}

		switch {
		case resp.StatusCode == http.StatusOK:
			jsonString, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if err != nil {
				return "", err
			}
			return string(jsonString), nil
		case resp.StatusCode == http.StatusTooManyRequests && attempt < RateLimitRetries:
			_ = resp.Body.Close()
			delay := RateLimitDelay(attempt, resp.Header)
			if c.onRateLimit != nil {
				c.onRateLimit(attempt+1, delay)
			}
			sleep(delay)
		default:
			apiErr := apierr.FromResponse(apierr.MarketOzon, resp)
			apiErr.Attempts = attempt + 1
			_ = resp.Body.Close()
			return "", apiErr
		}
	}
}

func (c Client) get(obj any, baseURL string, headers map[string]string, params map[string]string, body []byte) error {
	response, err := c.request(http.MethodGet, baseURL, headers, params, body)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(response), &obj)
}

func (c Client) post(baseURL string, headers map[string]string, params map[string]string, body []byte) (string, error) {
	response, err := c.request(http.MethodPost, baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

// v2PostingFboGet для получения ФБО заказа исходя из posting_number (нужен для извлечения товаров в возврате)
func (c Client) PostingFbo(postingNumber string) (PostingFBO, error) {
	baseURL := "https://api-seller.ozon.ru/v2/posting/fbo/get"

	body := []byte(`{
  		"posting_number": "` + postingNumber + `",
  		"translit": true,
  		"with": {
    		"analytics_data": false,
    		"financial_data": false
 		 }
	}`)

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var p PostingFBO
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal([]byte(response), &p)
	return p, err
}

// v3ReturnsCompanyFbo метод получения ФБО возвратов со статусом ReturnedToOzon. Получаем возвраты только тогда, когда возврат приедет на склад озон
func (c Client) v3ReturnsCompanyFbo(daysAgo, lastID int) (string, error) {
	baseURL := "https://api-seller.ozon.ru/v3/returns/company/fbo"

	body := []byte(`{
  		"filter": {
    		"status": [
      		"ReturnedToOzon"
    		]
  		},
  		"last_id":` + strconv.Itoa(lastID) + `,
  		"limit": 1000
	}`)

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, err
}
func (c Client) ReturnsList(lastID int, since, to string) (Returns, error) {
	baseURL := "https://api-seller.ozon.ru/v1/returns/list"

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

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var r Returns
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal([]byte(response), &r)
	return r, err
}

// v3ReturnsCompanyFbs метод получения ФБС возвратов со статусом moving_to_resale
func (c Client) v3ReturnsCompanyFbs(lastID int) (string, error) {
	baseURL := "https://api-seller.ozon.ru/v3/returns/company/fbs"

	body := []byte(`{
  		"filter": {
    		"status": "moving_to_resale"
  		},
  		"limit": 1000,
		"last_id": ` + strconv.Itoa(lastID) + `
	}`)

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}
	return response, err
}

// v3PostingFbsGet метод получения ФБС заказа исходя из posting_number (нужен для извлечения товаров в возврате)
func (c Client) PostingFbs(postingNumber string) (PostingFBS, error) {
	baseURL := "https://api-seller.ozon.ru/v3/posting/fbs/get"

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

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var p PostingFBS
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal([]byte(response), &p)
	return p, err
}

func (c Client) Stocks() (StocksList, error) {
	baseURL := "https://api-seller.ozon.ru/v2/analytics/stock_on_warehouses"

	body := []byte(`{
  		"limit": 1000,
  		"offset": 0,
  		"warehouse_type": "ALL"
	}`)

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var s StocksList
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal([]byte(response), &s)
	return s, err

}
func (c Client) Clusters() (ClustersList, error) {
	baseURL := "https://api-seller.ozon.ru/v1/cluster/list"

	body := []byte(`{
  		"cluster_type": "CLUSTER_TYPE_OZON"
	}`)

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var cl ClustersList
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return cl, err
	}
	err = json.Unmarshal([]byte(response), &cl)
	return cl, err
}

func (c Client) StocksAnalytics(skus []string) (StocksNew, error) {
	skusJSON, err := json.Marshal(skus)
	if err != nil {
		return StocksNew{}, err
	}

	baseURL := "https://api-seller.ozon.ru/v1/analytics/stocks"
	body := []byte(fmt.Sprintf(`{
  		"skus": %v
	}`, string(skusJSON)))

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var s StocksNew
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return s, err
	}
	err = json.Unmarshal([]byte(response), &s)
	return s, err

}

func (c Client) Warehouses(cursor string) (WarehouseList, error) {
	baseURL := "https://api-seller.ozon.ru/v2/warehouse/list"

	body := []byte(fmt.Sprintf(`{"cursor": "%v"}`, cursor))

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var w WarehouseList
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return w, err
	}
	err = json.Unmarshal([]byte(response), &w)
	return w, err
}

func (c Client) Products() (ProductList, error) {
	baseURL := "https://api-seller.ozon.ru/v3/product/list"
	body := []byte(`{
  "filter": {
    "visibility": "TO_SUPPLY"
  },
  "last_id": "",
  "limit": 1000
}`)

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var p ProductList
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal([]byte(response), &p)
	return p, err

}

func (c Client) ProductsWithAttributes() (ProductListWithAttributes, error) {
	baseURL := "https://api-seller.ozon.ru/v4/product/info/attributes"
	body := []byte(`{
  "filter": {
    "visibility": "VISIBLE"
  },
  "limit": 1000,
  "sort_dir": "ASC"
}`)

	headers := map[string]string{
		"Content-Type": "application/json",
		"Client-Id":    c.clientID,
		"Api-Key":      c.apiKey,
	}

	params := map[string]string{}

	var p ProductListWithAttributes
	response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal([]byte(response), &p)
	return p, err
}
