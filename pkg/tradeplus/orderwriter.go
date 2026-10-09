package tradeplus

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"tradebot/pkg/client/google"
	"tradebot/pkg/db"
)

// OrdersDaysAgo — отчёты по заказам и отгрузкам пишутся за вчерашний день.
const OrdersDaysAgo = 1

// MSK — дни отчётов считаются по Москве.
var MSK = time.FixedZone("MSK", 3*3600)

// OrdersReport — содержимое дневного листа «Заказы <маркетплейс>-<число>»:
// по секции на кабинет, секции идут друг под другом.
type OrdersReport struct {
	Day      time.Time
	Sections []OrdersSection
}

// OrdersSection — блоки одного кабинета. Блоки стоят через колонку: A:B, D:E, G:H.
type OrdersSection struct {
	Blocks []OrdersBlock
}

// OrdersBlock — строки «артикул | количество» под заголовком, например «Заказы FBO».
type OrdersBlock struct {
	Title  string
	Counts map[string]int
}

// OrdersSource собирает отчёт маркетплейса за сутки day (MSK).
type OrdersSource interface {
	OrdersReport(ctx context.Context, day time.Time) (OrdersReport, error)
}

// ordersSheets — операции с Google-таблицей, которые нужны OrdersWriter.
type ordersSheets interface {
	Spreadsheet(spreadsheetID string) (string, []google.SheetStat, error)
	BatchGet(spreadsheetID string, ranges []string) ([][][]interface{}, error)
	AddSheets(spreadsheetID string, titles []string) error
	BatchClear(spreadsheetID string, ranges []string) error
	BatchWrite(spreadsheetID string, data []google.ValueRange) error
}

// OrdersWriter ведёт ежедневные отчёты по заказам маркетплейса в таблице sheetLink.
// На каждое число месяца в таблице свой лист «<prefix><число>», в его A1 —
// «Отчет за 02.01.2006»: по этой отметке видно, за какую дату лист заполнен.
type OrdersWriter struct {
	spreadsheetID string
	prefix        string
	// dayInterval — пауза между запросами к маркетплейсу за разные дни.
	dayInterval time.Duration
	source      OrdersSource
	sheets      ordersSheets
}

func NewOrdersWriter(spreadsheetID, prefix string, dayInterval time.Duration, source OrdersSource) OrdersWriter {
	return OrdersWriter{
		spreadsheetID: spreadsheetID,
		prefix:        prefix,
		dayInterval:   dayInterval,
		source:        source,
		sheets:        NewSheetsService(),
	}
}

// NewSheetsService — клиент Google Sheets под аккаунтом бота.
func NewSheetsService() google.SheetsService {
	return google.NewSheetsService("pkg/client/google/token.json", "pkg/client/google/credentials.json")
}

// Sync заносит в таблицу дни из OrdersDays, которых в ней ещё нет. Возвращает
// записанные дни. Дни, которые не удалось собрать, попадают в ошибку и будут
// дописаны при следующем запуске.
func (w OrdersWriter) Sync(ctx context.Context, now time.Time) ([]time.Time, error) {
	if w.spreadsheetID == "" {
		return nil, errors.New("не задана таблица заказов")
	}

	_, sheetList, err := w.sheets.Spreadsheet(w.spreadsheetID)
	if err != nil {
		return nil, err
	}
	// Google не различает регистр в названиях листов: ищем без него, пишем в настоящее название.
	existing := make(map[string]string, len(sheetList))
	for _, s := range sheetList {
		existing[strings.ToLower(s.Title)] = s.Title
	}

	days, err := w.daysToWrite(OrdersDays(now), existing)
	if err != nil {
		return nil, err
	}

	var reports []OrdersReport
	var errs []error
	for i, day := range days {
		if i > 0 && w.dayInterval > 0 {
			if err := sleep(ctx, w.dayInterval); err != nil {
				errs = append(errs, err)
				break
			}
		}
		r, err := w.source.OrdersReport(ctx, day)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", day.Format("02.01"), err))
			continue
		}
		reports = append(reports, r)
	}

	if err := w.write(reports, existing); err != nil {
		return nil, errors.Join(append(errs, err)...)
	}

	written := make([]time.Time, 0, len(reports))
	for _, r := range reports {
		written = append(written, r.Day)
	}
	return written, errors.Join(errs...)
}

// daysToWrite оставляет дни, на листах которых нет отметки «Отчет за <дата>»:
// лист пустой, на нём прошлый месяц или его вовсе нет. Решение принимается только
// по таблице (A1 всех листов читаются одним запросом), к маркетплейсам бот идёт
// лишь за этими днями.
func (w OrdersWriter) daysToWrite(days []time.Time, existing map[string]string) ([]time.Time, error) {
	var ranges []string
	for _, d := range days {
		if title, ok := w.sheetTitle(d, existing); ok {
			ranges = append(ranges, google.SheetRange(title, "A1"))
		}
	}
	marks, err := w.sheets.BatchGet(w.spreadsheetID, ranges)
	if err != nil {
		return nil, fmt.Errorf("read report dates: %w", err)
	}

	var todo []time.Time
	i := 0
	for _, d := range days {
		if _, ok := w.sheetTitle(d, existing); !ok {
			todo = append(todo, d)
			continue
		}
		if i >= len(marks) || firstCell(marks[i]) != OrdersReportTitle(d) {
			todo = append(todo, d)
		}
		i++
	}
	return todo, nil
}

// write стирает старые данные на листах дней и записывает отчёты. Запросов два
// на все дни сразу, чтобы не упираться в лимит Google на запись.
func (w OrdersWriter) write(reports []OrdersReport, existing map[string]string) error {
	if len(reports) == 0 {
		return nil
	}

	var newSheets, toClear []string
	var data []google.ValueRange
	for _, r := range reports {
		title, ok := w.sheetTitle(r.Day, existing)
		if !ok {
			newSheets = append(newSheets, title)
		}
		toClear = append(toClear, r.clearRanges(title)...)
		data = append(data, r.valueRanges(title)...)
	}

	if err := w.sheets.AddSheets(w.spreadsheetID, newSheets); err != nil {
		return err
	}
	// A1 стирается вместе с данными: если запись ниже не пройдёт, у дня не будет
	// отметки, и он допишется при следующем запуске.
	if err := w.sheets.BatchClear(w.spreadsheetID, toClear); err != nil {
		return err
	}
	return w.sheets.BatchWrite(w.spreadsheetID, data)
}

// sheetTitle — лист дня day: его настоящее название в таблице и есть ли он там вообще.
func (w OrdersWriter) sheetTitle(day time.Time, existing map[string]string) (string, bool) {
	title := w.prefix + strconv.Itoa(day.Day())
	if actual, ok := existing[strings.ToLower(title)]; ok {
		return actual, true
	}
	return title, false
}

// OrdersDays — дни, которые проверяет Sync, по возрастанию: с 1-го числа текущего
// месяца по вчера. Первого числа это только вчерашний день — последний день
// прошлого месяца. Раньше 1-го числа не заглядываем: таблица заказов заводится
// на месяц, и в новой таблице листы с этими числами заполнились бы прошлым месяцем.
func OrdersDays(now time.Time) []time.Time {
	now = now.In(MSK)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, MSK)
	yesterday := today.AddDate(0, 0, -OrdersDaysAgo)

	from := today.AddDate(0, 0, 1-today.Day())
	if from.After(yesterday) {
		from = yesterday
	}

	var days []time.Time
	for d := from; !d.After(yesterday); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	return days
}

// OrdersReportTitle — отметка в A1 листа: за какую дату отчёт.
func OrdersReportTitle(day time.Time) string {
	return "Отчет за " + day.Format("02.01.2006")
}

// FormatDays — дни через запятую: «05.10, 06.10, 08.10».
func FormatDays(days []time.Time) string {
	s := make([]string, 0, len(days))
	for _, d := range days {
		s = append(s, d.Format("02.01"))
	}
	return strings.Join(s, ", ")
}

// valueRanges раскладывает отчёт по листу так же, как раньше: в A1 «Отчет за …»,
// со второй строки блоки через колонку (A:B, D:E, G:H) — заголовок блока, под
// ним «артикул | количество». Следующая секция (кабинет) начинается через
// пустую строку со своего «Отчет за …» в колонке A.
func (r OrdersReport) valueRanges(sheet string) []google.ValueRange {
	title := [][]interface{}{{OrdersReportTitle(r.Day)}}

	var out []google.ValueRange
	row := 1
	for _, s := range r.Sections {
		out = append(out, google.ValueRange{Range: google.SheetRange(sheet, "A"+strconv.Itoa(row)), Values: title})

		height := 0
		for i, b := range s.Blocks {
			values := b.values()
			from, to := blockColumns(i)
			out = append(out, google.ValueRange{
				Range:  google.SheetRange(sheet, fmt.Sprintf("%s%d:%s%d", from, row+1, to, row+len(values))),
				Values: values,
			})
			height = max(height, len(values))
		}
		row += height + 2
	}
	return out
}

// clearRanges — ячейки, которые заполняет бот: A1 и колонки блоков со второй строки до конца листа.
func (r OrdersReport) clearRanges(sheet string) []string {
	blocks := 0
	for _, s := range r.Sections {
		blocks = max(blocks, len(s.Blocks))
	}

	out := make([]string, 0, blocks+1)
	out = append(out, google.SheetRange(sheet, "A1"))
	for i := range blocks {
		from, to := blockColumns(i)
		out = append(out, google.SheetRange(sheet, from+"2:"+to))
	}
	return out
}

// values — строки блока: заголовок, затем артикулы по алфавиту.
func (b OrdersBlock) values() [][]interface{} {
	out := make([][]interface{}, 0, len(b.Counts)+1)
	out = append(out, []interface{}{b.Title})
	for _, article := range slices.Sorted(maps.Keys(b.Counts)) {
		out = append(out, []interface{}{article, b.Counts[article]})
	}
	return out
}

// OrdersCabinets — кабинеты, которые попадают в отчёт по заказам: включённые,
// по возрастанию ID. Порядок важен: по нему идут секции на листе Ozon, а без
// сортировки база отдаёт строки как придётся.
func (cc Cabinets) OrdersCabinets() Cabinets {
	out := make(Cabinets, 0, len(cc))
	for _, c := range cc {
		if c.StatusID == db.StatusEnabled {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(a, b Cabinet) int { return a.ID - b.ID })
	return out
}

// OrdersSpreadsheetID — таблица заказов маркетплейса. Она общая для всех его
// кабинетов, берётся первая заданная.
func (cc Cabinets) OrdersSpreadsheetID() string {
	if ids := cc.OrdersSpreadsheetIDs(); len(ids) > 0 {
		return ids[0]
	}
	return ""
}

// OrdersSpreadsheetIDs — разные таблицы заказов, заданные у кабинетов, в порядке
// кабинетов. В sheetLink может лежать как ID таблицы, так и ссылка на неё.
func (cc Cabinets) OrdersSpreadsheetIDs() []string {
	var ids []string
	for _, c := range cc {
		if c.SheetLink == nil {
			continue
		}
		if id, ok := google.ParseSpreadsheetID(*c.SheetLink); ok && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}

// blockColumns — колонки i-го блока секции: A:B, D:E, G:H, …
func blockColumns(i int) (string, string) {
	c := 'A' + rune(3*i)
	return string(c), string(c + 1)
}

func firstCell(rows [][]interface{}) string {
	if len(rows) == 0 || len(rows[0]) == 0 {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(rows[0][0]))
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
