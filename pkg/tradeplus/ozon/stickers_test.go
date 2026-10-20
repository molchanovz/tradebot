package ozon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"tradebot/pkg/client/ozon"

	"github.com/gen2brain/go-fitz"
	"github.com/jung-kurt/gofpdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pdfPages создаёт PDF, где на каждой странице напечатан свой текст.
func pdfPages(t *testing.T, pages ...string) []byte {
	t.Helper()
	pdf := gofpdf.New("P", "mm", "A6", "")
	pdf.SetFont("Helvetica", "", 14)
	for _, text := range pages {
		pdf.AddPage()
		pdf.Text(10, 20, text)
	}
	var buf bytes.Buffer
	require.NoError(t, pdf.Output(&buf))
	return buf.Bytes()
}

func openPDF(t *testing.T, data []byte) *fitz.Document {
	t.Helper()
	doc, err := fitz.NewFromMemory(data)
	require.NoError(t, err)
	t.Cleanup(func() { _ = doc.Close() })
	return doc
}

// fakeOzon — тестовый сервер вместо Озона для заданий на этикетки. labelsFile собирает
// PDF для отправлений задания; statuses — статусы, которые задание проходит до completed.
type fakeOzon struct {
	t          *testing.T
	labelsFile func(postings []string) []byte
	statuses   []string
	unprinted  []ozon.UnprintedPosting

	mu      sync.Mutex
	created [][]string // posting_numbers каждого созданного задания
	polls   int
}

func (f *fakeOzon) client() ozon.Client {
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	f.t.Cleanup(srv.Close)
	return ozon.NewClient("id", "key").WithBaseURL(srv.URL)
}

func (f *fakeOzon) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch {
	case r.URL.Path == "/v3/posting/fbs/package-label/create":
		var req struct {
			PostingNumbers []string `json:"posting_numbers"`
		}
		assert.NoError(f.t, json.NewDecoder(r.Body).Decode(&req))
		f.created = append(f.created, req.PostingNumbers)
		id := len(f.created) - 1
		_, _ = fmt.Fprintf(w, `{"tasks":[{"task_id":%d,"task_type":"small_label"},{"task_id":%d,"task_type":"big_label"}]}`, 1000+id, id)

	case r.URL.Path == "/v2/posting/fbs/package-label/get":
		var req struct {
			TaskID int `json:"task_id"`
		}
		assert.NoError(f.t, json.NewDecoder(r.Body).Decode(&req))
		assert.Less(f.t, req.TaskID, 1000, "нужно задание на обычную этикетку (big_label)")

		var info ozon.LabelTaskInfo
		info.Status.Code = ozon.LabelTaskCompleted
		if f.polls < len(f.statuses) {
			info.Status.Code = f.statuses[f.polls]
		}
		f.polls++
		if info.Status.Code == ozon.LabelTaskCompleted {
			info.FileURL = fmt.Sprintf("http://%s/file/%d", r.Host, req.TaskID)
			info.Status.UnprintedPostings = f.unprinted
			info.Status.PostingsCount = len(f.created[req.TaskID])
		}
		assert.NoError(f.t, json.NewEncoder(w).Encode(info))

	case strings.HasPrefix(r.URL.Path, "/file/"):
		id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/file/"))
		_, _ = w.Write(f.labelsFile(f.created[id]))

	default:
		f.t.Errorf("unexpected path %s", r.URL.Path)
	}
}

// fbs — отправления с номерами numbers без кода ScanIt.
func fbs(numbers ...string) []ozon.FbsPosting {
	postings := make([]ozon.FbsPosting, len(numbers))
	for i, n := range numbers {
		postings[i].PostingNumber = n
	}
	return postings
}

func noSleep(t *testing.T) {
	sleep = func(time.Duration) {}
	t.Cleanup(func() { sleep = time.Sleep })
}

func TestOpenLabels(t *testing.T) {
	noSleep(t)
	numbers := []string{"111-0001-1", "111-0001-12", "222-0002-1"}
	postings := fbs(numbers...)

	t.Run("one file, pages found by posting number", func(t *testing.T) {
		fake := &fakeOzon{t: t, statuses: []string{ozon.LabelTaskPending, ozon.LabelTaskInProgress}, labelsFile: func(p []string) []byte {
			// Озон может вернуть этикетки не в порядке запроса
			texts := make([]string, 0, len(p))
			for _, posting := range p {
				texts = append(texts, "Posting "+posting+" FBS")
			}
			slices.Reverse(texts)
			return pdfPages(t, texts...)
		}}

		labels, err := openLabels(fake.client(), postings)
		require.NoError(t, err)
		defer labels.Close()

		assert.False(t, labels.oneByOne())
		assert.Equal(t, map[string]int{"111-0001-1": 2, "111-0001-12": 1, "222-0002-1": 0}, labels.pages)
		assert.Equal(t, [][]string{numbers}, fake.created, "одно задание на всю пачку")
		assert.Equal(t, 3, fake.polls, "ждали, пока задание не станет completed")
	})

	t.Run("no posting numbers in file: labels one by one", func(t *testing.T) {
		fake := &fakeOzon{t: t, labelsFile: func(p []string) []byte {
			return pdfPages(t, make([]string, len(p))...) // страницы без текста
		}}

		labels, err := openLabels(fake.client(), postings)
		require.NoError(t, err)
		defer labels.Close()
		assert.True(t, labels.oneByOne())

		img, err := labels.image("111-0001-12")
		require.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, [][]string{numbers, {"111-0001-12"}}, fake.created,
			"по порядку страниц не сопоставляем — заказываем этикетку отдельно")
	})

	t.Run("unprinted postings stop the print", func(t *testing.T) {
		fake := &fakeOzon{t: t,
			unprinted:  []ozon.UnprintedPosting{{PostingNumber: "222-0002-1", Message: "отправление отменено"}},
			labelsFile: func(p []string) []byte { return pdfPages(t, p...) },
		}

		_, err := openLabels(fake.client(), postings)
		require.Error(t, err)
		assert.Equal(t, "озон не сформировал этикетки для 1 из 3 отправлений: 222-0002-1 (отправление отменено)", err.Error())
	})

	t.Run("task error", func(t *testing.T) {
		fake := &fakeOzon{t: t, statuses: []string{ozon.LabelTaskError}}

		_, err := openLabels(fake.client(), postings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "озон не смог сформировать этикетки")
	})
}

func TestLabelPages(t *testing.T) {
	t.Run("by posting number on the page", func(t *testing.T) {
		// -1 и -12 — разные отправления
		doc := openPDF(t, pdfPages(t, "Posting 111-0001-12 FBS", "Posting 111-0001-1 FBS", "Posting 222-0002-1"))

		pages, err := labelPages(doc, fbs("111-0001-1", "111-0001-12", "222-0002-1"))
		require.NoError(t, err)
		assert.Equal(t, map[string]int{"111-0001-1": 1, "111-0001-12": 0, "222-0002-1": 2}, pages)
	})

	t.Run("by ScanIt code printed in parts", func(t *testing.T) {
		// новая этикетка: номера отправления нет, код ScanIt напечатан как «ii5008470 9305»
		doc := openPDF(t, pdfPages(t, "OZON ii5008471 0001", "OZON ii5008470 9305"))
		postings := fbs("111-0001-1", "222-0002-1")
		postings[0].Scanit = "ii50084709305"
		postings[1].Scanit = "ii50084710001"

		pages, err := labelPages(doc, postings)
		require.NoError(t, err)
		assert.Equal(t, map[string]int{"111-0001-1": 1, "222-0002-1": 0}, pages)
	})

	t.Run("no numbers at all", func(t *testing.T) {
		doc := openPDF(t, pdfPages(t, "", ""))
		postings := fbs("111-0001-1", "222-0002-1")
		postings[0].Scanit = "ii50084709305"

		_, err := labelPages(doc, postings)
		require.ErrorIs(t, err, errNoPostingNumbers)
	})

	t.Run("some postings missing", func(t *testing.T) {
		doc := openPDF(t, pdfPages(t, "Posting 111-0001-1", "Posting 333-0003-1"))

		_, err := labelPages(doc, fbs("111-0001-1", "222-0002-1"))
		require.Error(t, err)
		require.NotErrorIs(t, err, errNoPostingNumbers)
		assert.Contains(t, err.Error(), "222-0002-1")
	})

	t.Run("two postings on one page", func(t *testing.T) {
		doc := openPDF(t, pdfPages(t, "Posting 111-0001-1 and 222-0002-1", "Posting 222-0002-1"))

		_, err := labelPages(doc, fbs("111-0001-1", "222-0002-1"))
		require.Error(t, err)
		assert.Equal(t, "на одной странице файла этикеток Озона (1) нашлись отправления 111-0001-1 и 222-0002-1", err.Error())
	})
}

func TestContainsCode(t *testing.T) {
	assert.True(t, containsCode("Отправление 111-0001-1 FBS", "111-0001-1", true))
	assert.True(t, containsCode("111-0001-1", "111-0001-1", true))
	assert.False(t, containsCode("111-0001-12", "111-0001-1", true))
	assert.False(t, containsCode("9111-0001-1", "111-0001-1", true))
	assert.True(t, containsCode("111-0001-12 и 111-0001-1", "111-0001-1", true))

	// в тексте без пробелов код ScanIt приклеен к соседним словам
	assert.True(t, containsCode("OZONii50084709305Артикул", "ii50084709305", false))
	assert.False(t, containsCode("OZONii500847093051", "ii50084709305", false))
}

func TestPickLabelTask(t *testing.T) {
	task, ok := pickLabelTask([]ozon.LabelTask{{TaskID: 1, TaskType: "small_label"}, {TaskID: 2, TaskType: "big_label"}})
	assert.True(t, ok)
	assert.Equal(t, int64(2), task.TaskID)

	task, ok = pickLabelTask([]ozon.LabelTask{{TaskID: 3, TaskType: "small_label"}})
	assert.True(t, ok)
	assert.Equal(t, int64(3), task.TaskID)

	_, ok = pickLabelTask(nil)
	assert.False(t, ok)
}

func TestLabelLayout(t *testing.T) {
	// новая этикетка Озона горизонтальная — без поворота, на ширину области
	w, h, rotate := labelLayout(1.85)
	assert.False(t, rotate)
	assert.InDelta(t, labelAreaW, w, 0.01)
	assert.InDelta(t, labelAreaW/1.85, h, 0.01)

	// вертикальная этикетка крупнее, если её повернуть
	w, h, rotate = labelLayout(0.6)
	assert.True(t, rotate)
	assert.InDelta(t, labelAreaW, h, 0.01, "после поворота высота картинки — ширина стикера")
	assert.InDelta(t, labelAreaW*0.6, w, 0.01)
	assert.LessOrEqual(t, w, labelAreaH)

	// очень широкая упирается в ширину, а не в высоту
	w, h, rotate = labelLayout(4)
	assert.False(t, rotate)
	assert.InDelta(t, labelAreaW, w, 0.01)
	assert.InDelta(t, labelAreaW/4, h, 0.01)
}

func TestStickerLayout(t *testing.T) {
	inPage := func(t *testing.T, r rect, rotated bool) {
		t.Helper()
		w, h := r.w, r.h
		if rotated {
			w, h = h, w
		}
		cx, cy := r.x+r.w/2, r.y+r.h/2
		assert.GreaterOrEqual(t, cx-w/2, 0.0)
		assert.LessOrEqual(t, cx+w/2, stickerW+1e-9)
		assert.GreaterOrEqual(t, cy-h/2, stickerMinMargin-1e-9)
		assert.LessOrEqual(t, cy+h/2, stickerH-stickerMinMargin+1e-9)
	}

	t.Run("horizontal label: barcode full width, composition centered", func(t *testing.T) {
		label, rotate, barcode := stickerLayout(1.85, 1.5)
		assert.False(t, rotate)
		inPage(t, label, false)
		inPage(t, barcode, false)

		assert.InDelta(t, labelAreaW, barcode.w, 0.01, "баркод той же ширины, что область этикетки")
		assert.InDelta(t, labelAreaW/1.5, barcode.h, 0.01, "в своих пропорциях")
		assert.InDelta(t, label.y+label.h+labelBarcodeGap, barcode.y, 0.01, "баркод сразу под этикеткой")
		assert.InDelta(t, label.y, stickerH-(barcode.y+barcode.h), 0.01, "поля сверху и снизу равны")
	})

	t.Run("tall rotated label shrinks the barcode to fit", func(t *testing.T) {
		label, rotate, barcode := stickerLayout(0.6, 1.5)
		assert.True(t, rotate)
		inPage(t, label, true)
		inPage(t, barcode, false)

		labelBottom := label.y + label.h/2 + label.w/2 // после поворота видимая высота — ширина картинки
		assert.InDelta(t, labelBottom+labelBarcodeGap, barcode.y, 0.01)
		assert.InDelta(t, 1.5, barcode.w/barcode.h, 0.01)
	})
}
