package ozon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"tradebot/pkg/client/apierr"
)

const (
	// postingsPageLimit — максимум отправлений на страницу в /v4/posting/fbs/list и /v3/posting/fbo/list.
	postingsPageLimit = 100
	// maxPostingsPages — предохранитель от бесконечной пагинации.
	maxPostingsPages = 1000

	// LabelsBatchLimit — сколько отправлений Озон принимает в одном задании на этикетки.
	LabelsBatchLimit = 1000

	// downloadTimeout — файл с этикетками на сотни отправлений качается дольше обычного запроса.
	downloadTimeout = 2 * time.Minute

	headerClientID = "Client-Id"
	headerAPIKey   = "Api-Key"
)

// PostingProduct — товар в отправлении.
type PostingProduct struct {
	OfferID  string `json:"offer_id"`
	Name     string `json:"name"`
	Sku      int64  `json:"sku"`
	Quantity int    `json:"quantity"`
	Price    struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"price"`
}

// FbsPosting — FBS-отправление из /v4/posting/fbs/list. Только поля, которые нужны боту.
type FbsPosting struct {
	PostingNumber string `json:"posting_number"`
	OrderID       int64  `json:"order_id"`
	OrderNumber   string `json:"order_number"`
	Status        string `json:"status"`
	Substatus     string `json:"substatus"`
	// Scanit — код ScanIt, напечатанный на новой этикетке вместо номера отправления
	// (например, ii30525117479; последние цифры на этикетке выделены отдельно).
	Scanit         string `json:"scanit"`
	DeliveryMethod struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		WarehouseID int64  `json:"warehouse_id"`
		Warehouse   string `json:"warehouse"`
	} `json:"delivery_method"`
	Products      []PostingProduct `json:"products"`
	FinancialData struct {
		ClusterFrom string `json:"cluster_from"`
		ClusterTo   string `json:"cluster_to"`
	} `json:"financial_data"`
}

// UnitArticles — артикул каждой единицы товара в отправлении по порядку:
// A ×2 и B ×1 дают A, A, B.
func (p FbsPosting) UnitArticles() []string {
	var articles []string
	for _, product := range p.Products {
		for range product.Quantity {
			articles = append(articles, product.OfferID)
		}
	}
	return articles
}

// FirstArticle — артикул первого товара отправления или пустая строка, если товаров нет.
func (p FbsPosting) FirstArticle() string {
	if len(p.Products) == 0 {
		return ""
	}
	return p.Products[0].OfferID
}

// PostingslistFbs — FBS-отправления со всех страниц /v4/posting/fbs/list.
type PostingslistFbs struct {
	Result struct {
		PostingsFBS []FbsPosting
	}
}

// FboPosting — FBO-отправление из /v3/posting/fbo/list. Только поля, которые нужны боту.
type FboPosting struct {
	PostingNumber string           `json:"posting_number"`
	Status        string           `json:"status"`
	Products      []PostingProduct `json:"products"`
	FinancialData struct {
		ClusterTo string `json:"cluster_to"`
	} `json:"financial_data"`
}

// PostingslistFbo — FBO-отправления со всех страниц /v3/posting/fbo/list.
type PostingslistFbo struct {
	Result []FboPosting
}

// FbsFilter — фильтр списка FBS-отправлений. Since и To — время в RFC 3339, обязательны;
// пустые Statuses и WarehouseIDs не фильтруют.
type FbsFilter struct {
	Since, To    string
	Statuses     []string
	WarehouseIDs []int64
}

type postingsFilter struct {
	Since        string   `json:"since"`
	To           string   `json:"to"`
	Statuses     []string `json:"statuses,omitempty"`
	WarehouseIDs []string `json:"warehouse_ids,omitempty"`
}

type postingsListRequest struct {
	Cursor  string         `json:"cursor,omitempty"`
	Filter  postingsFilter `json:"filter"`
	Limit   int            `json:"limit"`
	SortDir string         `json:"sort_dir"`
	With    struct {
		FinancialData bool `json:"financial_data"`
	} `json:"with"`
}

type postingsPage[T any] struct {
	Cursor   string `json:"cursor"`
	HasNext  bool   `json:"has_next"`
	Postings []T    `json:"postings"`
}

// PostingsListFbs возвращает все FBS-отправления по фильтру, проходя страницы /v4/posting/fbs/list.
func (c Client) PostingsListFbs(filter FbsFilter) (PostingslistFbs, error) {
	f := postingsFilter{Since: filter.Since, To: filter.To, Statuses: filter.Statuses}
	for _, id := range filter.WarehouseIDs {
		f.WarehouseIDs = append(f.WarehouseIDs, strconv.FormatInt(id, 10))
	}

	var list PostingslistFbs
	postings, err := listPostings[FbsPosting](c, "/v4/posting/fbs/list", f)
	list.Result.PostingsFBS = postings
	return list, err
}

// PostingsListFbo возвращает все FBO-отправления за период, проходя страницы /v3/posting/fbo/list.
func (c Client) PostingsListFbo(since, to string) (PostingslistFbo, error) {
	postings, err := listPostings[FboPosting](c, "/v3/posting/fbo/list", postingsFilter{Since: since, To: to})
	return PostingslistFbo{Result: postings}, err
}

// listPostings проходит все страницы метода списка отправлений по cursor.
func listPostings[T any](c Client, path string, filter postingsFilter) ([]T, error) {
	req := postingsListRequest{Filter: filter, Limit: postingsPageLimit, SortDir: "ASC"}
	req.With.FinancialData = true

	var all []T
	for page := 1; page <= maxPostingsPages; page++ {
		var resp postingsPage[T]
		if err := c.postJSON(path, req, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Postings...)

		if !resp.HasNext {
			return all, nil
		}
		if resp.Cursor == "" || resp.Cursor == req.Cursor {
			return nil, fmt.Errorf("ozon %s: has_next без нового cursor на странице %d", path, page)
		}
		req.Cursor = resp.Cursor
	}
	return nil, fmt.Errorf("ozon %s: больше %d страниц", path, maxPostingsPages)
}

// LabelTask — задание Озона на формирование файла с этикетками.
type LabelTask struct {
	TaskID   int64  `json:"task_id"`
	TaskType string `json:"task_type"` // big_label — обычная этикетка, small_label — маленькая
}

// CreateLabelTasks создаёт задания на этикетки для отправлений в статусе awaiting_deliver
// (не больше LabelsBatchLimit за раз): /v3/posting/fbs/package-label/create.
func (c Client) CreateLabelTasks(postingNumbers []string) ([]LabelTask, error) {
	var resp struct {
		Tasks []LabelTask `json:"tasks"`
	}
	err := c.postJSON("/v3/posting/fbs/package-label/create", map[string][]string{"posting_numbers": postingNumbers}, &resp)
	return resp.Tasks, err
}

// Статусы задания на этикетки.
const (
	LabelTaskPending    = "pending"
	LabelTaskInProgress = "in_progress"
	LabelTaskCompleted  = "completed"
	LabelTaskError      = "error"
)

// LabelTaskInfo — состояние задания на этикетки.
type LabelTaskInfo struct {
	FileURL string `json:"file_url"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Status struct {
		Code                 string             `json:"code"`
		PostingsCount        int                `json:"postings_count"`
		PrintedPostingsCount int                `json:"printed_postings_count"`
		UnprintedPostings    []UnprintedPosting `json:"unprinted_postings"`
	} `json:"status"`
}

// UnprintedPosting — отправление, для которого Озон не сформировал этикетку.
type UnprintedPosting struct {
	PostingNumber string `json:"posting_number"`
	Message       string `json:"message"`
}

// LabelTask возвращает состояние задания на этикетки: /v2/posting/fbs/package-label/get.
func (c Client) LabelTask(taskID int64) (LabelTaskInfo, error) {
	var info LabelTaskInfo
	err := c.postJSON("/v2/posting/fbs/package-label/get", map[string]int64{"task_id": taskID}, &info)
	return info, err
}

// DownloadFile скачивает файл по ссылке из ответа Озона. Ключи кабинета отправляются
// только на домены Озона.
func (c Client) DownloadFile(fileURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	if isOzonHost(req.URL) {
		req.Header.Set(headerClientID, c.clientID)
		req.Header.Set(headerAPIKey, c.apiKey)
	}

	hc := *c.hc
	hc.Timeout = downloadTimeout
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка скачивания файла: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apierr.FromResponse(apierr.MarketOzon, resp)
	}
	return io.ReadAll(resp.Body)
}

func isOzonHost(u *url.URL) bool {
	host := u.Hostname()
	for _, domain := range []string{"ozon.ru", "ozone.ru"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

// postJSON отправляет body в метод Seller API path и разбирает ответ в out.
func (c Client) postJSON(path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"Content-Type": "application/json", //nolint:goconst // как в остальных методах клиента
		headerClientID: c.clientID,
		headerAPIKey:   c.apiKey,
	}
	response, err := c.post(c.baseURL+path, headers, nil, b)
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(response), out); err != nil {
		return fmt.Errorf("ozon %s: разбор ответа: %w", path, err)
	}
	return nil
}
