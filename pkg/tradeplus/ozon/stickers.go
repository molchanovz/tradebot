package ozon

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // пропорции картинок баркодов в imageAspect
	"log"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"time"

	"tradebot/pkg/client/ozon"
	"tradebot/pkg/tradeplus"

	"github.com/fogleman/gg"
	"github.com/gen2brain/go-fitz"
	"github.com/jung-kurt/gofpdf"
)

type StickerManager struct {
	clientID, token string
	printedOrders   map[string]struct{}
	warehouseID     int64
}

func NewStickerManager(clientID, token string, printedOrders map[string]struct{}, warehouseID int64) StickerManager {
	return StickerManager{
		clientID:      clientID,
		token:         token,
		printedOrders: printedOrders,
		warehouseID:   warehouseID,
	}
}

const (
	AllLabels = "all"
	NewLabels = "new"
)

func (m StickerManager) GetAllLabels(progressChan chan tradeplus.Progress) ([]string, error) {
	tradeplus.CreateDirectories()

	log.Printf("ozonStickers: clientID=%s warehouse=%d GetAllLabels: start", m.clientID, m.warehouseID)

	progress := tradeplus.NewProgressReporter(progressChan)
	orderIds, err := m.getSortedFbsOrders(progress)
	if err != nil {
		log.Printf("ozonStickers: clientID=%s warehouse=%d GetAllLabels: getSortedFbsOrders err=%v", m.clientID, m.warehouseID, err)
		return []string{}, err
	}

	log.Printf("ozonStickers: clientID=%s warehouse=%d GetAllLabels: будет напечатано %d заказов", m.clientID, m.warehouseID, len(orderIds.Result.PostingsFBS))

	readyPdfPath, err := m.getReadyPdf(orderIds, progress)
	if err != nil {
		log.Printf("ozonStickers: clientID=%s warehouse=%d GetAllLabels: getReadyPdf err=%v", m.clientID, m.warehouseID, err)
		return []string{}, err
	}

	return readyPdfPath, nil
}

func (m StickerManager) GetNewLabels(progressChan chan tradeplus.Progress) ([]string, ozon.PostingslistFbs, error) {
	tradeplus.CreateDirectories()

	log.Printf("ozonStickers: clientID=%s warehouse=%d GetNewLabels: start printedOrders=%d", m.clientID, m.warehouseID, len(m.printedOrders))

	progress := tradeplus.NewProgressReporter(progressChan)
	orders, err := m.getSortedFbsOrders(progress)
	if err != nil {
		log.Printf("ozonStickers: clientID=%s warehouse=%d GetNewLabels: getSortedFbsOrders err=%v", m.clientID, m.warehouseID, err)
		return []string{}, ozon.PostingslistFbs{}, err
	}

	//Проверка, есть ли новые заказы
	newOrders := ozon.PostingslistFbs{}
	var skippedAsPrinted []string
	for _, p := range orders.Result.PostingsFBS {
		if _, ok := m.printedOrders[p.PostingNumber]; !ok {
			newOrders.Result.PostingsFBS = append(newOrders.Result.PostingsFBS, p)
		} else {
			skippedAsPrinted = append(skippedAsPrinted, p.PostingNumber)
		}
	}

	log.Printf("ozonStickers: clientID=%s warehouse=%d GetNewLabels: fetched=%d printedSkip=%d new=%d skippedSample=%v",
		m.clientID, m.warehouseID, len(orders.Result.PostingsFBS), len(skippedAsPrinted), len(newOrders.Result.PostingsFBS), firstN(skippedAsPrinted, 5))

	if len(newOrders.Result.PostingsFBS) == 0 {
		log.Printf("ozonStickers: clientID=%s warehouse=%d GetNewLabels: ErrNoRows", m.clientID, m.warehouseID)
		return []string{}, newOrders, ErrNoRows
	}

	log.Printf("ozonStickers: clientID=%s warehouse=%d GetNewLabels: будет напечатано %d новых заказов", m.clientID, m.warehouseID, len(newOrders.Result.PostingsFBS))

	readyPdfPaths, err := m.getReadyPdf(newOrders, progress)
	if err != nil {
		return []string{}, newOrders, err
	}

	return readyPdfPaths, newOrders, nil
}

func firstN(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// client — клиент Озона, который пишет в прогресс о паузах перед повтором после 429.
func (m StickerManager) client(progress *tradeplus.ProgressReporter) ozon.Client {
	return ozon.NewClient(m.clientID, m.token).WithRateLimitHook(func(retry int, delay time.Duration) {
		seconds := max(1, int(delay.Round(time.Second)/time.Second))
		progress.Note(fmt.Sprintf("Озон ограничил частоту запросов, повтор %d из %d через %d с…", retry, ozon.RateLimitRetries, seconds))
	})
}

const (
	// stickersBatch — сколько отправлений идёт в одно задание Озона на этикетки и в один
	// итоговый PDF.
	stickersBatch = 200
	// labelsPollInterval и labelsWaitTimeout — как часто и сколько ждём файл с этикетками.
	labelsPollInterval = 2 * time.Second
	labelsWaitTimeout  = 3 * time.Minute
)

// sleep подменяется в тестах, чтобы не ждать между опросами задания.
var sleep = time.Sleep

// errNoPostingNumbers — в файле этикеток не нашёлся ни один номер или код ScanIt
// отправления (например, этикетки вставлены картинкой), и страницы не сопоставить
// с отправлениями.
var errNoPostingNumbers = errors.New("в файле этикеток Озона не читаются номера отправлений")

// getReadyPdf печатает стикеры пачками по stickersBatch: находит этикетку каждого
// отправления и добавляет к ней баркод артикула. Возвращает по одному PDF на пачку.
func (m StickerManager) getReadyPdf(orders ozon.PostingslistFbs, progress *tradeplus.ProgressReporter) ([]string, error) {
	tradeplus.CreateDirectories()

	client := m.client(progress)
	postings := orders.Result.PostingsFBS
	total := len(postings)

	var resultFiles []string
	for start := 0; start < total; start += stickersBatch {
		batch := postings[start:min(start+stickersBatch, total)]
		batchNum := start/stickersBatch + 1

		progress.Stage("Запрашиваю этикетки у Озона", start, total)
		labels, err := openLabels(client, batch)
		if err != nil {
			return nil, err
		}
		if labels.oneByOne() {
			log.Printf("ozonStickers: clientID=%s batch=%d: %v — запрашиваю этикетки по одной; текст первой страницы: %q",
				m.clientID, batchNum, errNoPostingNumbers, labels.firstPageText())
		}

		readyPdfPath, err := buildStickers(batch, labels, batchNum, progress, start, total)
		labels.Close()
		if err != nil {
			return nil, err
		}
		resultFiles = append(resultFiles, readyPdfPath)
	}

	if len(resultFiles) == 0 {
		return nil, errors.New("не было создано ни одного PDF файла")
	}
	return resultFiles, nil
}

// labelSource отдаёт этикетку каждого отправления пачки. Обычно все этикетки лежат
// в одном файле и страница отправления находится по его номеру. Если номера в файле
// не читаются, этикетки заказываются по одной: сопоставлять страницы по порядку нельзя,
// иначе этикетка может достаться чужому баркоду.
type labelSource struct {
	client ozon.Client
	doc    *fitz.Document
	pages  map[string]int // страница отправления в doc; nil — этикетки по одной
}

// openLabels заказывает у Озона один файл с этикетками на все postings.
func openLabels(client ozon.Client, postings []ozon.FbsPosting) (*labelSource, error) {
	doc, err := fetchLabels(client, postingNumbers(postings))
	if err != nil {
		return nil, err
	}

	pages, err := labelPages(doc, postings)
	if err != nil && !errors.Is(err, errNoPostingNumbers) {
		_ = doc.Close()
		return nil, err
	}
	return &labelSource{client: client, doc: doc, pages: pages}, nil
}

func (s *labelSource) oneByOne() bool { return s.pages == nil }

// image — картинка этикетки отправления posting.
func (s *labelSource) image(posting string) (image.Image, error) {
	if !s.oneByOne() {
		return s.doc.Image(s.pages[posting])
	}

	doc, err := fetchLabels(s.client, []string{posting})
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	if doc.NumPage() == 0 {
		return nil, fmt.Errorf("озон прислал пустой файл этикетки %s", posting)
	}
	// в файле на одно отправление его этикетка — первая страница
	return doc.Image(0)
}

func (s *labelSource) Close() { _ = s.doc.Close() }

// firstPageText — начало текста первой страницы файла, для лога.
func (s *labelSource) firstPageText() string {
	if s.doc.NumPage() == 0 {
		return ""
	}
	text, err := s.doc.Text(0)
	if err != nil {
		return "ошибка: " + err.Error()
	}
	text = strings.Join(strings.Fields(text), " ")
	if r := []rune(text); len(r) > 300 {
		text = string(r[:300]) + "…"
	}
	return text
}

// fetchLabels заказывает у Озона файл с этикетками на postings, ждёт его и открывает.
func fetchLabels(client ozon.Client, postings []string) (*fitz.Document, error) {
	tasks, err := client.CreateLabelTasks(postings)
	if err != nil {
		return nil, fmt.Errorf("создание задания на этикетки: %w", err)
	}
	task, ok := pickLabelTask(tasks)
	if !ok {
		return nil, errors.New("озон не вернул задание на этикетки")
	}

	info, err := waitLabels(client, task.TaskID)
	if err != nil {
		return nil, err
	}

	data, err := client.DownloadFile(info.FileURL)
	if err != nil {
		return nil, fmt.Errorf("скачивание этикеток: %w", err)
	}
	doc, err := fitz.NewFromMemory(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия PDF этикеток: %w", err)
	}
	return doc, nil
}

// pickLabelTask выбирает задание на обычную этикетку, если Озон вернул несколько.
func pickLabelTask(tasks []ozon.LabelTask) (ozon.LabelTask, bool) {
	for _, t := range tasks {
		if t.TaskType == "big_label" {
			return t, true
		}
	}
	if len(tasks) > 0 {
		return tasks[0], true
	}
	return ozon.LabelTask{}, false
}

// waitLabels опрашивает задание, пока Озон не сформирует файл. Если хоть одна этикетка
// не сформирована, возвращает ошибку: иначе такие отправления отметились бы напечатанными.
func waitLabels(client ozon.Client, taskID int64) (ozon.LabelTaskInfo, error) {
	deadline := time.Now().Add(labelsWaitTimeout)
	for {
		info, err := client.LabelTask(taskID)
		if err != nil {
			return info, fmt.Errorf("статус задания на этикетки: %w", err)
		}

		switch info.Status.Code {
		case ozon.LabelTaskCompleted:
			if len(info.Status.UnprintedPostings) > 0 {
				return info, unprintedError(info)
			}
			if info.FileURL == "" {
				return info, errors.New("озон сформировал этикетки, но не прислал ссылку на файл")
			}
			return info, nil
		case ozon.LabelTaskError:
			return info, fmt.Errorf("озон не смог сформировать этикетки: %s %s", info.Error.Code, info.Error.Message)
		}

		if time.Now().After(deadline) {
			return info, fmt.Errorf("озон не подготовил этикетки за %v (статус задания: %s)", labelsWaitTimeout, info.Status.Code)
		}
		sleep(labelsPollInterval)
	}
}

func unprintedError(info ozon.LabelTaskInfo) error {
	unprinted := info.Status.UnprintedPostings
	parts := make([]string, 0, min(len(unprinted), 5))
	for _, u := range unprinted[:min(len(unprinted), 5)] {
		parts = append(parts, fmt.Sprintf("%s (%s)", u.PostingNumber, u.Message))
	}
	if len(unprinted) > len(parts) {
		parts = append(parts, fmt.Sprintf("и ещё %d", len(unprinted)-len(parts)))
	}
	return fmt.Errorf("озон не сформировал этикетки для %d из %d отправлений: %s",
		len(unprinted), info.Status.PostingsCount, strings.Join(parts, ", "))
}

// buildStickers собирает итоговый PDF пачки: стикер на каждую единицу товара —
// этикетка Озона её отправления и баркод артикула. У отправления с несколькими
// единицами этикетка повторяется на каждом его стикере.
func buildStickers(batch []ozon.FbsPosting, labels *labelSource, batchNum int, progress *tradeplus.ProgressReporter, done, total int) (string, error) {
	// по одной этикетки идут медленно — показываем каждую
	stage, every := "Готовлю стикеры", 5
	if labels.oneByOne() {
		stage, every = "Номера на этикетках не читаются, запрашиваю их по одной", 1
	}

	combinedPDFs := make([]string, 0, len(batch))
	for i, posting := range batch {
		articles := posting.UnitArticles()
		if len(articles) == 0 {
			return "", fmt.Errorf("в отправлении %s нет товаров", posting.PostingNumber)
		}

		label, err := labels.image(posting.PostingNumber)
		if err != nil {
			return "", fmt.Errorf("этикетка %s: %w", posting.PostingNumber, err)
		}

		for unit, article := range articles {
			finalPDF := fmt.Sprintf("%s%s_%d.pdf", tradeplus.ReadyPath, posting.PostingNumber, unit+1)
			if err = combineLabelWithBarcode(label, finalPDF, article); err != nil {
				return "", fmt.Errorf("ошибка объединения PDF с баркодом: %w", err)
			}
			combinedPDFs = append(combinedPDFs, finalPDF)
		}

		if n := done + i + 1; n%every == 0 && n != total {
			progress.Stage(stage, n, total)
		}
	}

	progress.Stage("Собираю PDF", done+len(batch), total)
	readyPdfPath := fmt.Sprintf("%s/ozon_%d.pdf", tradeplus.BatchesPath, batchNum)
	if err := mergePDFsInDirectory(combinedPDFs, readyPdfPath); err != nil {
		return "", fmt.Errorf("ошибка объединения PDF для батча %d: %w", batchNum, err)
	}
	if !fileExists(readyPdfPath) {
		return "", fmt.Errorf("итоговый PDF для батча %d не создан", batchNum)
	}
	return readyPdfPath, nil
}

func postingNumbers(postings []ozon.FbsPosting) []string {
	numbers := make([]string, len(postings))
	for i, p := range postings {
		numbers[i] = p.PostingNumber
	}
	return numbers
}

type pageTexts interface {
	NumPage() int
	Text(pageNumber int) (string, error)
}

// labelPages находит для каждого отправления страницу его этикетки в файле Озона:
// на этикетке напечатан номер отправления (старый формат) или его код ScanIt (новый).
// Если не нашлось ни одно отправление, возвращает errNoPostingNumbers; если нашлись
// не все или два отправления оказались на одной странице — ошибку: такой файл печатать
// нельзя.
func labelPages(doc pageTexts, postings []ozon.FbsPosting) (map[string]int, error) {
	pageTexts := make([]labelText, doc.NumPage())
	for i := range pageTexts {
		text, err := doc.Text(i)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения текста этикетки, страница %d: %w", i+1, err)
		}
		pageTexts[i] = newLabelText(text)
	}

	pages := make(map[string]int, len(postings))
	owners := make(map[int]string, len(postings))
	var missing []string
	for _, posting := range postings {
		page := slices.IndexFunc(pageTexts, func(text labelText) bool { return text.has(posting) })
		if page < 0 {
			missing = append(missing, posting.PostingNumber)
			continue
		}
		if other, taken := owners[page]; taken {
			return nil, fmt.Errorf("на одной странице файла этикеток Озона (%d) нашлись отправления %s и %s", page+1, other, posting.PostingNumber)
		}
		owners[page] = posting.PostingNumber
		pages[posting.PostingNumber] = page
	}

	switch {
	case len(missing) == 0:
		return pages, nil
	case len(missing) == len(postings):
		return nil, errNoPostingNumbers
	default:
		return nil, fmt.Errorf("в файле этикеток Озона (%d стр.) не нашлись отправления: %s",
			len(pageTexts), strings.Join(firstN(missing, 5), ", "))
	}
}

// labelText — текст страницы этикетки в двух видах: с пробелами между словами и совсем
// без пробелов (код ScanIt на этикетке печатается частями: ii5008470 9305).
type labelText struct {
	spaced, compact string
}

func newLabelText(text string) labelText {
	words := strings.Fields(text)
	return labelText{spaced: strings.Join(words, " "), compact: strings.Join(words, "")}
}

// has — напечатан ли на странице номер отправления или его код ScanIt.
func (t labelText) has(posting ozon.FbsPosting) bool {
	if containsCode(t.spaced, posting.PostingNumber, true) {
		return true
	}
	scanit := strings.Join(strings.Fields(posting.Scanit), "")
	return scanit != "" && containsCode(t.compact, scanit, false)
}

// containsCode — есть ли в text код целиком: цифра сразу после кода значит другой,
// более длинный код (123-0001-1 не должен совпасть с 123-0001-12). Цифру перед кодом
// проверяем, только если checkBefore: в тексте без пробелов перед кодом ScanIt может
// оказаться соседнее слово.
func containsCode(text, code string, checkBefore bool) bool {
	isDigit := func(b byte) bool { return b >= '0' && b <= '9' }
	for from := 0; ; {
		i := strings.Index(text[from:], code)
		if i < 0 {
			return false
		}
		start, end := from+i, from+i+len(code)
		if (!checkBefore || start == 0 || !isDigit(text[start-1])) && (end == len(text) || !isDigit(text[end])) {
			return true
		}
		from = start + 1
	}
}

func (m StickerManager) getSortedFbsOrders(progress *tradeplus.ProgressReporter) (ozon.PostingslistFbs, error) {
	filter := ozon.FbsFilter{
		Since:    time.Now().AddDate(0, 0, -7).Format("2006-01-02T15:04:05.000Z"),
		To:       time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05.000Z"),
		Statuses: []string{"awaiting_deliver"},
	}
	if m.warehouseID != 0 {
		filter.WarehouseIDs = []int64{m.warehouseID}
	}

	log.Printf("ozonStickers: clientID=%s warehouse=%d fetch awaiting_deliver since=%s to=%s", m.clientID, m.warehouseID, filter.Since, filter.To)

	progress.Stage("Получаю список заказов из Озона", 0, 0)
	orders, err := m.client(progress).PostingsListFbs(filter)
	if err != nil {
		log.Printf("ozonStickers: clientID=%s warehouse=%d PostingsListFbs err=%v", m.clientID, m.warehouseID, err)
		return orders, err
	}

	whDist := map[int64]int{}
	for _, p := range orders.Result.PostingsFBS {
		whDist[p.DeliveryMethod.WarehouseID]++
	}
	log.Printf("ozonStickers: clientID=%s warehouse=%d fetched total=%d byWarehouse=%v", m.clientID, m.warehouseID, len(orders.Result.PostingsFBS), whDist)

	// Склад фильтрует и Озон (warehouse_ids), здесь — страховка: на чужой склад стикер печатать нельзя.
	if m.warehouseID != 0 {
		before := len(orders.Result.PostingsFBS)
		filtered := orders.Result.PostingsFBS[:0]
		for _, p := range orders.Result.PostingsFBS {
			if p.DeliveryMethod.WarehouseID == m.warehouseID {
				filtered = append(filtered, p)
			}
		}
		orders.Result.PostingsFBS = filtered
		log.Printf("ozonStickers: clientID=%s warehouse=%d warehouse-filter before=%d after=%d", m.clientID, m.warehouseID, before, len(orders.Result.PostingsFBS))
	}

	if len(orders.Result.PostingsFBS) == 0 {
		log.Printf("ozonStickers: clientID=%s warehouse=%d getSortedFbsOrders: ErrNoRows", m.clientID, m.warehouseID)
		return orders, ErrNoRows
	}

	sort.Slice(orders.Result.PostingsFBS, func(i, j int) bool {
		return orders.Result.PostingsFBS[i].FirstArticle() < orders.Result.PostingsFBS[j].FirstArticle()
	})
	return orders, nil
}

// Раскладка стикера 75×120 мм: сверху этикетка Озона (не больше labelAreaW×labelAreaH),
// под ней через labelBarcodeGap баркод артикула той же ширины, вся композиция —
// по центру стикера по высоте, с полями не меньше stickerMinMargin.
const (
	stickerW         = 75.0
	stickerH         = 120.0
	labelAreaW       = 65.0
	labelAreaH       = 55.0
	labelBarcodeGap  = 6.0
	stickerMinMargin = 3.0

	// defaultBarcodeAspect — пропорции баркода, если картинку не удалось прочитать
	// (картинки в assets/barcodes — 673×449).
	defaultBarcodeAspect = 1.5
)

// rect — положение картинки на стикере, мм: левый верхний угол и размер до поворота.
type rect struct{ x, y, w, h float64 }

// stickerLayout раскладывает стикер для этикетки и баркода с пропорциями labelAspect
// и barcodeAspect (ширина / высота). rotate — повернуть этикетку на 90° вокруг её центра.
func stickerLayout(labelAspect, barcodeAspect float64) (label rect, rotate bool, barcode rect) {
	labelW, labelH, rotate := labelLayout(labelAspect)
	visualH := labelH
	if rotate {
		visualH = labelW
	}

	barcodeW, barcodeH := labelAreaW, labelAreaW/barcodeAspect
	if maxH := stickerH - 2*stickerMinMargin - visualH - labelBarcodeGap; barcodeH > maxH {
		barcodeW, barcodeH = maxH*barcodeAspect, maxH
	}

	top := (stickerH - (visualH + labelBarcodeGap + barcodeH)) / 2
	centerY := top + visualH/2
	label = rect{x: (stickerW - labelW) / 2, y: centerY - labelH/2, w: labelW, h: labelH}
	barcode = rect{x: (stickerW - barcodeW) / 2, y: top + visualH + labelBarcodeGap, w: barcodeW, h: barcodeH}
	return label, rotate, barcode
}

// imageAspect — пропорции картинки в файле path (ширина / высота).
func imageAspect(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, err
	}
	if cfg.Height == 0 {
		return 0, fmt.Errorf("%s: нулевая высота", path)
	}
	return float64(cfg.Width) / float64(cfg.Height), nil
}

// labelLayout вписывает этикетку с пропорциями aspect (ширина / высота) в область под неё
// без поворота или повёрнутой на 90° — как выйдет крупнее. Новая горизонтальная этикетка
// идёт без поворота на ширину labelAreaW. w и h — размер картинки до поворота, мм.
func labelLayout(aspect float64) (w, h float64, rotate bool) {
	fit := func(boxW, boxH float64) (float64, float64) {
		w, h := boxW, boxW/aspect
		if h > boxH {
			w, h = boxH*aspect, boxH
		}
		return w, h
	}

	uprightW, uprightH := fit(labelAreaW, labelAreaH)
	// повёрнутая картинка ложится шириной на высоту области
	rotatedW, rotatedH := fit(labelAreaH, labelAreaW)
	if rotatedW*rotatedH > uprightW*uprightH {
		return rotatedW, rotatedH, true
	}
	return uprightW, uprightH, false
}

// combineLabelWithBarcode собирает стикер 75×120 мм: сверху этикетка Озона img
// (повёрнутая, только если так она крупнее), снизу баркод артикула — см. stickerLayout.
func combineLabelWithBarcode(img image.Image, outputPath, article string) error {
	tmpImg := outputPath + ".jpg"

	defer os.Remove(tmpImg)

	file, err := os.Create(tmpImg)
	if err != nil {
		return fmt.Errorf("ошибка создания JPEG: %w", err)
	}
	defer file.Close()

	if err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("ошибка сохранения JPEG: %w", err)
	}

	// Баркод артикула
	barcodePath := tradeplus.BarcodesPath + article + ".png"
	if !fileExists(barcodePath) {
		barcodePath = tradeplus.GeneratedPath + article + "_generated.png"
		if err = createBarcodeWithSKU(article, barcodePath, 40); err != nil {
			log.Println(err)
			barcodePath = tradeplus.BarcodesPath + "0.png"
		}
	}
	barcodeAspect, err := imageAspect(barcodePath)
	if err != nil {
		log.Printf("ozonStickers: пропорции баркода %s: %v", barcodePath, err)
		barcodeAspect = defaultBarcodeAspect
	}

	bounds := img.Bounds()
	label, rotate, barcode := stickerLayout(float64(bounds.Dx())/float64(bounds.Dy()), barcodeAspect)

	pdf := gofpdf.New("P", "mm", "", "")
	pdf.AddPageFormat("P", gofpdf.SizeType{Wd: stickerW, Ht: stickerH})

	if rotate {
		pdf.TransformBegin()
		pdf.TransformRotate(90, label.x+label.w/2, label.y+label.h/2)
	}
	pdf.ImageOptions(tmpImg, label.x, label.y, label.w, label.h, false,
		gofpdf.ImageOptions{ImageType: "JPG"}, 0, "")
	if rotate {
		pdf.TransformEnd()
	}

	pdf.ImageOptions(barcodePath, barcode.x, barcode.y, barcode.w, barcode.h, false,
		gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	return pdf.OutputFileAndClose(outputPath)
}

// Функция для создания изображения с текстом (артикул товара) и сохранения в PNG
func createBarcodeWithSKU(sku string, outputPath string, fontSize float64) error {
	const imgWidth = 580  // Ширина изображения в пикселях
	const imgHeight = 400 // Высота изображения в пикселях

	// Создание нового изображения
	dc := gg.NewContext(imgWidth, imgHeight)

	// Установка фона (белый цвет)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Загрузка шрифта и установка его размера
	if err := dc.LoadFontFace(tradeplus.FontPath, fontSize); err != nil {
		return err
	}

	// Настройка текста
	dc.SetRGB(0, 0, 0) // Цвет текста (черный)
	text := sku

	// Добавление текста на изображение
	dc.DrawStringAnchored(text, float64(imgWidth)/2, float64(imgHeight)/2, 0.5, 0.5)

	// Сохранение изображения в формате PNG
	return dc.SavePNG(outputPath)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func mergePDFsInDirectory(orderSlice []string, outputFile string) error {
	// Проверяем, есть ли PDF-файлы для объединения
	if len(orderSlice) == 0 {
		return errors.New("нет PDF-файлов в директории")
	}

	// Формируем команду для выполнения merge PDF через pdfcpu
	args := append([]string{"merge", outputFile}, orderSlice...)
	cmd := exec.Command("pdfcpu", args...)

	// Запуск команды
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка выполнения pdfcpu: %v, %s", err, string(output))
	}
	return nil
}
