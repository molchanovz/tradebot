package wb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"tradebot/pkg/client/apierr"
)

type Client struct {
	hc           *http.Client
	token        string
	analyticsURL string
}

func NewClient(token string) Client {
	return Client{
		hc: &http.Client{
			Timeout: time.Second * 10,
		},
		token:        token,
		analyticsURL: "https://seller-analytics-api.wildberries.ru",
	}
}

// WithAnalyticsURL возвращает копию клиента, которая ходит на analyticsURL вместо
// seller-analytics-api. Нужен тестам, которые поднимают свой сервер вместо WB.
func (c Client) WithAnalyticsURL(analyticsURL string) Client {
	c.analyticsURL = analyticsURL
	return c
}

func (c Client) request(reqType, baseURL string, headers, params map[string]string, body []byte) (int, string, error) {
	req, err := http.NewRequest(reqType, baseURL, bytes.NewBuffer(body))

	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("ошибка создания запроса: %w", err)
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
		return http.StatusInternalServerError, "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	jsonString, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, "", err
	}
	return resp.StatusCode, string(jsonString), nil
}

func (c Client) get(baseURL string, headers map[string]string, params map[string]string, body []byte) (int, string, error) {
	return c.request(http.MethodGet, baseURL, headers, params, body)
}

func (c Client) post(baseURL string, headers map[string]string, params map[string]string, body []byte) (int, string, error) {
	return c.request(http.MethodPost, baseURL, headers, params, body)
}

func (c Client) stocksFbo() (string, error) {
	baseURL := "https://statistics-api.wildberries.ru/api/v1/supplier/stocks"

	body := []byte(``)

	headers := map[string]string{
		"Authorization": c.token,
	}

	params := map[string]string{
		"dateFrom": "2019-06-20",
	}

	_, response, err := c.get(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c Client) getOrdersBySupplyID(supplyID string) (string, error) {
	baseURL := "https://marketplace-api.wildberries.ru/api/marketplace/v3/supplies/" + supplyID + "/order-ids"
	body := []byte(``)

	headers := map[string]string{
		"Authorization": c.token,
	}

	_, response, err := c.get(baseURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}

// returnsPage запрашивает страницу отчёта о возвратах товаров продавцу за [dateFrom, dateTo]
// (не больше 31 дня). На 204 (за период ничего нет) возвращает nil.
func (c Client) returnsPage(ctx context.Context, dateFrom, dateTo time.Time, status string, limit, offset int) ([]byte, error) {
	params := url.Values{}
	params.Set("dateFrom", dateFrom.Format(time.DateOnly))
	params.Set("dateTo", dateTo.Format(time.DateOnly))
	params.Set("status", status)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))

	return c.getRetrying(ctx, returnsPath, c.analyticsURL+returnsPath+"?"+params.Encode())
}

func (c Client) getCodesByOrderID(orderID int) (string, error) {
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
		Orders: []int{orderID},
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": c.token,
	}

	_, response, err := c.post(fullURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}

// GetOrdersFBS отдает фбс заказы
func (c Client) GetOrdersFBS(dateFrom, dateTo int) (*OrdersListFBS, error) {
	baseURL := "https://marketplace-api.wildberries.ru/api/v3/orders"
	body := []byte(``)

	headers := map[string]string{
		"Authorization": c.token,
	}

	var all OrdersListFBS
	next := 0
	for {
		params := map[string]string{
			"limit":    "1000",
			"next":     strconv.Itoa(next),
			"dateFrom": strconv.Itoa(dateFrom),
			"dateTo":   strconv.Itoa(dateTo),
		}

		_, response, err := c.get(baseURL, headers, params, body)
		if err != nil {
			return nil, err
		}

		var page OrdersListFBS
		if err := json.Unmarshal([]byte(response), &page); err != nil {
			return nil, err
		}

		all.OrdersFBS = append(all.OrdersFBS, page.OrdersFBS...)
		all.Next = page.Next

		if page.Next == 0 || len(page.OrdersFBS) == 0 {
			break
		}
		next = page.Next
	}

	return &all, nil
}

func (c Client) ordersFBSStatus(orderID int) (string, error) {
	baseURL := "https://marketplace-api.wildberries.ru/api/v3/orders/status"

	type RequestBody struct {
		Orders []int `json:"orders"`
	}

	data := RequestBody{
		Orders: []int{orderID},
	}

	// Преобразование данных в JSON
	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("ошибка при преобразовании данных в JSON: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": c.token,
	}

	_, response, err := c.post(baseURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}

	return response, nil
}
func (c Client) getCards(nmID *int, updatedAt *time.Time, limit *int) (string, error) {
	baseURL := "https://content-api.wildberries.ru/content/v2/get/cards/list"

	type RequestBody struct {
		Settings struct {
			Sort struct {
				Ascending bool `json:"ascending"`
			} `json:"sort"`
			Cursor struct {
				UpdatedAt *time.Time `json:"updatedAt"`
				NmId      *int       `json:"nmID"`
				Limit     *int       `json:"limit"`
			} `json:"cursor"`
			Filter struct {
				WithPhoto int `json:"withPhoto"`
			} `json:"filter"`
		} `json:"settings"`
	}

	var data RequestBody
	data.Settings.Cursor.Limit = limit
	data.Settings.Cursor.NmId = nmID
	data.Settings.Cursor.UpdatedAt = updatedAt

	data.Settings.Sort.Ascending = true
	data.Settings.Filter.WithPhoto = -1

	if limit != nil {
		data.Settings.Cursor.Limit = limit
	}

	// Преобразование данных в JSON
	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("ошибка при преобразовании данных в JSON: %w", err)
	}

	headers := map[string]string{
		"Authorization": c.token,
	}

	status, response, err := c.post(baseURL, headers, make(map[string]string), body)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", &apierr.Error{Market: apierr.MarketWB, Method: "/content/v2/get/cards/list", Status: status, Body: response}
	}

	return response, nil
}

/*
API метод для получения всех заказов за дату, равную (now() - daysAgo). Максимум 1 запрос в минуту
*/
func (c Client) apiOrdersALL(daysAgo, flag int) (string, error) {
	date := time.Now().AddDate(0, 0, -daysAgo)

	baseURL := "https://statistics-api.wildberries.ru/api/v1/supplier/orders"

	body := []byte(``)

	headers := map[string]string{
		"Authorization": c.token,
	}

	params := map[string]string{
		"dateFrom": date.Format("2006-01-02"),
		"flag":     strconv.Itoa(flag),
	}

	_, response, err := c.get(baseURL, headers, params, body)
	if err != nil {
		return "", err
	}

	return response, nil
}

const (
	statisticsURL = "https://statistics-api.wildberries.ru"
	returnsPath   = "/api/analytics/v1/item-returns"
	// rateLimitRetries — сколько раз повторяем запрос после 429.
	rateLimitRetries = 3
	// rateLimitRetryDefault — пауза после 429, если WB не прислал X-Ratelimit-Retry:
	// методы статистики и отчёт о возвратах отдают не больше одного запроса в минуту.
	rateLimitRetryDefault = time.Minute
	// slowTimeout — статистика и отчёты WB отвечают медленнее остальных методов клиента (у них 10 секунд).
	slowTimeout = time.Minute
)

// statisticsOnDate запрашивает метод statistics-api с flag=1 — все записи с датой
// date (время значения не имеет).
func (c Client) statisticsOnDate(ctx context.Context, path string, date time.Time) ([]byte, error) {
	params := url.Values{}
	params.Set("dateFrom", date.Format("2006-01-02"))
	params.Set("flag", "1")
	return c.getRetrying(ctx, path, statisticsURL+path+"?"+params.Encode())
}

// getRetrying делает GET u. На 429 ждёт столько, сколько просит WB, и повторяет.
// 200 и 204 (нет данных) — успех, иначе *apierr.Error с path в Method.
func (c Client) getRetrying(ctx context.Context, path, u string) ([]byte, error) {
	hc := &http.Client{Timeout: slowTimeout}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания запроса: %w", err)
		}
		req.Header.Set("Authorization", c.token)

		resp, err := hc.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		switch {
		case resp.StatusCode == http.StatusOK:
			return body, nil
		case resp.StatusCode == http.StatusNoContent:
			return nil, nil
		case resp.StatusCode == http.StatusTooManyRequests && attempt < rateLimitRetries:
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryAfter(resp.Header)):
			}
		default:
			return nil, &apierr.Error{Market: apierr.MarketWB, Method: path, Status: resp.StatusCode, Body: string(body)}
		}
	}
}

// retryAfter — сколько ждать после 429: WB пишет это в X-Ratelimit-Retry (секунды).
func retryAfter(h http.Header) time.Duration {
	if s, err := strconv.Atoi(h.Get("X-Ratelimit-Retry")); err == nil && s > 0 {
		return time.Duration(s)*time.Second + time.Second
	}
	return rateLimitRetryDefault
}

// reviewPageSize is the WB feedbacks page size. Reviews() pages through all
// unanswered feedbacks; reviewMaxFetch caps the total to avoid runaway loops.
const (
	reviewPageSize = 100
	reviewMaxFetch = 10000
)

// Reviews returns ALL unanswered feedbacks, paging through the WB API.
// The WB endpoint returns at most reviewPageSize per request, so a single
// take=100/skip=0 call (the previous behaviour) only ever saw the first 100
// reviews and never reached the rest.
func (c Client) Reviews() (*Review, error) {
	var result *Review

	for skip := 0; skip < reviewMaxFetch; skip += reviewPageSize {
		page, err := c.reviewsPage(skip, reviewPageSize)
		if err != nil {
			return nil, err
		}

		if result == nil {
			result = page
		} else {
			result.Data.Feedbacks = append(result.Data.Feedbacks, page.Data.Feedbacks...)
		}

		// Last page reached once WB returns fewer than a full page.
		if len(page.Data.Feedbacks) < reviewPageSize {
			break
		}
	}

	return result, nil
}

// reviewsPage fetches a single page of unanswered feedbacks.
func (c Client) reviewsPage(skip, take int) (*Review, error) {
	baseURL := "https://feedbacks-api.wildberries.ru/api/v1/feedbacks"

	body := []byte(``)

	headers := map[string]string{
		"Authorization": c.token,
	}

	params := map[string]string{
		"isAnswered": "false",
		"take":       strconv.Itoa(take),
		"skip":       strconv.Itoa(skip),
	}

	_, response, err := c.get(baseURL, headers, params, body)
	if err != nil {
		return nil, err
	}

	var review *Review

	err = json.Unmarshal([]byte(response), &review)
	if err != nil {
		return nil, err
	} else if review.Error {
		return nil, errors.New(response)
	}

	return review, nil
}

func (c Client) AnswerReview(id, answer string) error {
	baseURL := "https://feedbacks-api.wildberries.ru/api/v1/feedbacks/answer"

	type RequestBody struct {
		Id   string `json:"id"`
		Text string `json:"text"`
	}

	var data = RequestBody{
		Id:   id,
		Text: answer,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("data marshal failed: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": c.token,
	}

	params := map[string]string{}

	status, response, err := c.post(baseURL, headers, params, body)
	if err != nil {
		return fmt.Errorf("response get failed: %w", err)
	}

	if status == http.StatusNoContent {
		return nil
	}

	var errResp struct {
		Title     string `json:"title"`
		Detail    string `json:"detail"`
		Code      string `json:"code"`
		Error     bool   `json:"error"`
		ErrorText string `json:"errorText"`
	}
	if jsonErr := json.Unmarshal([]byte(response), &errResp); jsonErr == nil {
		if errResp.ErrorText != "" {
			return fmt.Errorf("wb answer review failed (status %d): %s", status, errResp.ErrorText)
		}
		if errResp.Detail != "" {
			return fmt.Errorf("wb answer review failed (status %d): %s", status, errResp.Detail)
		}
		if errResp.Title != "" {
			return fmt.Errorf("wb answer review failed (status %d): %s", status, errResp.Title)
		}
	}

	if response != "" {
		return fmt.Errorf("wb answer review failed (status %d): %s", status, response)
	}
	return fmt.Errorf("wb answer review failed (status %d)", status)
}

func GetUnix(date time.Time) int64 {
	nowStr := fmt.Sprint(date.Format("2006-01-02"), "T21:00:00")
	t, _ := time.Parse("2006-01-02T15:04:05", nowStr)
	return t.Unix()
}
