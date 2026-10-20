package ozon

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testClient — клиент, который ходит в тестовый сервер вместо Seller API.
func testClient(t *testing.T, h http.HandlerFunc) Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := NewClient("id", "key")
	c.baseURL = srv.URL
	return c
}

func TestPostingsListFbs(t *testing.T) {
	var requests []postingsListRequest
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v4/posting/fbs/list", r.URL.Path)
		assert.Equal(t, "id", r.Header.Get("Client-Id"))

		var req postingsListRequest
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&req)) {
			return
		}
		requests = append(requests, req)

		// две страницы: первая с курсором, вторая последняя
		if req.Cursor == "" {
			_, _ = io.WriteString(w, `{"cursor":"page2","has_next":true,"postings":[
				{"posting_number":"1-0001-1","status":"awaiting_deliver","scanit":"ii30525117479",
				 "delivery_method":{"warehouse_id":42},
				 "products":[{"offer_id":"A","quantity":2,"price":{"amount":"100.00","currency":"RUB"}}],
				 "financial_data":{"cluster_to":"Москва"}}]}`)
			return
		}
		_, _ = io.WriteString(w, `{"cursor":"","has_next":false,"postings":[{"posting_number":"2-0001-1","products":[{"offer_id":"B","quantity":1}]}]}`)
	})

	list, err := c.PostingsListFbs(FbsFilter{Since: "2026-10-13T00:00:00Z", To: "2026-10-21T00:00:00Z", Statuses: []string{"awaiting_deliver"}, WarehouseIDs: []int64{42}})
	require.NoError(t, err)

	postings := list.Result.PostingsFBS
	require.Len(t, postings, 2)
	assert.Equal(t, "1-0001-1", postings[0].PostingNumber)
	assert.Equal(t, "ii30525117479", postings[0].Scanit)
	assert.Equal(t, int64(42), postings[0].DeliveryMethod.WarehouseID)
	assert.Equal(t, "A", postings[0].Products[0].OfferID)
	assert.Equal(t, 2, postings[0].Products[0].Quantity)
	assert.Equal(t, "100.00", postings[0].Products[0].Price.Amount)
	assert.Equal(t, "Москва", postings[0].FinancialData.ClusterTo)
	assert.Equal(t, "2-0001-1", postings[1].PostingNumber)

	require.Len(t, requests, 2)
	assert.Empty(t, requests[0].Cursor)
	assert.Equal(t, "page2", requests[1].Cursor)
	for _, req := range requests {
		assert.Equal(t, postingsPageLimit, req.Limit)
		assert.Equal(t, []string{"awaiting_deliver"}, req.Filter.Statuses)
		assert.Equal(t, []string{"42"}, req.Filter.WarehouseIDs)
		assert.True(t, req.With.FinancialData)
	}
}

func TestPostingsListStuckCursor(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"cursor":"same","has_next":true,"postings":[]}`)
	})

	_, err := c.PostingsListFbo("2026-10-13T00:00:00Z", "2026-10-14T00:00:00Z")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "без нового cursor")
}

func TestPostingsListFbo(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/posting/fbo/list", r.URL.Path)
		var req map[string]any
		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&req)) {
			return
		}
		assert.NotContains(t, req["filter"], "statuses", "пустой фильтр статусов не отправляем")

		_, _ = io.WriteString(w, `{"has_next":false,"postings":[{"posting_number":"3-0001-1","status":"delivered",
			"products":[{"offer_id":"C","quantity":3,"price":{"amount":"5","currency":"RUB"}}],"financial_data":{"cluster_to":"Казань"}}]}`)
	})

	list, err := c.PostingsListFbo("2026-10-13T00:00:00Z", "2026-10-14T00:00:00Z")
	require.NoError(t, err)
	require.Len(t, list.Result, 1)
	assert.Equal(t, "delivered", list.Result[0].Status)
	assert.Equal(t, 3, list.Result[0].Products[0].Quantity)
	assert.Equal(t, "Казань", list.Result[0].FinancialData.ClusterTo)
}

func TestLabelTasks(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.URL.Path {
		case "/v3/posting/fbs/package-label/create":
			assert.JSONEq(t, `{"posting_numbers":["1-0001-1","2-0001-1"]}`, string(body))
			_, _ = io.WriteString(w, `{"tasks":[{"task_id":7,"task_type":"small_label"},{"task_id":8,"task_type":"big_label"}]}`)
		case "/v2/posting/fbs/package-label/get":
			assert.JSONEq(t, `{"task_id":8}`, string(body))
			_, _ = io.WriteString(w, `{"file_url":"https://cdn1.ozone.ru/labels.pdf","status":{"code":"completed","postings_count":2,"printed_postings_count":1,
				"unprinted_postings":[{"posting_number":"2-0001-1","message":"отменено"}]}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	tasks, err := c.CreateLabelTasks([]string{"1-0001-1", "2-0001-1"})
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	assert.Equal(t, LabelTask{TaskID: 8, TaskType: "big_label"}, tasks[1])

	info, err := c.LabelTask(8)
	require.NoError(t, err)
	assert.Equal(t, LabelTaskCompleted, info.Status.Code)
	assert.Equal(t, "https://cdn1.ozone.ru/labels.pdf", info.FileURL)
	require.Len(t, info.Status.UnprintedPostings, 1)
	assert.Equal(t, "2-0001-1", info.Status.UnprintedPostings[0].PostingNumber)
}

func TestDownloadFile(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Api-Key")
		_, _ = io.WriteString(w, "%PDF-1.4")
	}))
	defer srv.Close()

	data, err := NewClient("id", "key").DownloadFile(srv.URL + "/labels.pdf")
	require.NoError(t, err)
	assert.Equal(t, "%PDF-1.4", string(data))
	assert.Empty(t, gotKey, "ключ кабинета уходит только на домены Озона")
}

func TestIsOzonHost(t *testing.T) {
	for raw, want := range map[string]bool{
		"https://cdn1.ozone.ru/a.pdf":       true,
		"https://api-seller.ozon.ru/a.pdf":  true,
		"https://ozon.ru/a.pdf":             true,
		"https://evilozon.ru/a.pdf":         false,
		"https://ozon.ru.example.com/a.pdf": false,
	} {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.Equal(t, want, isOzonHost(u), raw)
	}
}

func TestFbsPostingArticles(t *testing.T) {
	var p FbsPosting
	assert.Empty(t, p.UnitArticles())
	assert.Empty(t, p.FirstArticle())

	p.Products = []PostingProduct{{OfferID: "A", Quantity: 2}, {OfferID: "B", Quantity: 1}}
	assert.Equal(t, []string{"A", "A", "B"}, p.UnitArticles())
	assert.Equal(t, "A", p.FirstArticle())
}
