package tradeplus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"tradebot/pkg/client/google"
)

// ShipmentsBlock — отгрузки одного кабинета за день: блок на листе «Отправлено на … FBS-<число>».
type ShipmentsBlock struct {
	Cabinet string
	Header  []string
	Rows    [][]string
}

// ShipmentsSource собирает отгрузки кабинета за сутки day (MSK).
type ShipmentsSource interface {
	ShipmentsBlock(ctx context.Context, day time.Time) (ShipmentsBlock, error)
}

// shipmentsSheets — операции с Google-таблицей, которые нужны ShipmentsWriter.
type shipmentsSheets interface {
	Spreadsheet(spreadsheetID string) (string, []google.SheetStat, error)
	BatchGet(spreadsheetID string, ranges []string) ([][][]interface{}, error)
	AddSheets(spreadsheetID string, titles []string) error
	BatchWriteUserEntered(spreadsheetID string, data []google.ValueRange) error
}

// ShipmentsWriter заполняет дневные листы отгрузок маркетплейса в складской таблице
// (settings.shipmentsSheetId). Лист дня — блоки кабинетов подряд: строка с названием
// кабинета, шапка, строки отгрузок. На этот формат завязаны формулы складской таблицы
// (шапку ищут во второй строке), поэтому меток на листе нет: заполненным считается
// любой непустой лист, и его бот не трогает.
type ShipmentsWriter struct {
	// Progress, если задан, вызывается перед загрузкой (written пуст) и после
	// каждого записанного дня: сколько дней уже внесено из total.
	Progress func(written []time.Time, total int)

	spreadsheetID string
	prefix        string
	sources       []ShipmentsSource
	sheets        shipmentsSheets
}

func NewShipmentsWriter(spreadsheetID, prefix string, sources []ShipmentsSource) ShipmentsWriter {
	return ShipmentsWriter{
		spreadsheetID: spreadsheetID,
		prefix:        prefix,
		sources:       sources,
		sheets:        NewSheetsService(),
	}
}

// NewShipmentsWriters собирает писателей отгрузок маркетплейса: по одному на каждую
// таблицу отгрузок, заданную у кабинетов (обычно она одна на все кабинеты).
// Кабинеты без таблицы отгрузок пропускаются.
func NewShipmentsWriters(ctx context.Context, cabinets Cabinets, prefix string, source func(context.Context, Cabinet) (ShipmentsSource, error)) ([]ShipmentsWriter, error) {
	var ids []string
	byID := map[string][]ShipmentsSource{}
	for _, c := range cabinets {
		id, ok := google.ParseSpreadsheetID(c.Settings.ShipmentsSheetID)
		if !ok {
			continue
		}
		s, err := source(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("cabinet %d: %w", c.ID, err)
		}
		if _, seen := byID[id]; !seen {
			ids = append(ids, id)
		}
		byID[id] = append(byID[id], s)
	}

	writers := make([]ShipmentsWriter, 0, len(ids))
	for _, id := range ids {
		writers = append(writers, NewShipmentsWriter(id, prefix, byID[id]))
	}
	return writers, nil
}

// Sync заполняет листы дней из OrdersDays, которых в таблице нет или которые пустые.
// Данные запрашиваются по одному дню, каждый день пишется сразу. Возвращает
// записанные дни; дни, которые не удалось собрать, попадают в ошибку и будут
// дописаны при следующем запуске.
func (w ShipmentsWriter) Sync(ctx context.Context, now time.Time) ([]time.Time, error) {
	if w.spreadsheetID == "" {
		return nil, errors.New("не задана таблица отгрузок")
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

	all := OrdersDays(now)
	yesterday := all[len(all)-1]
	days, err := w.emptyDays(all, existing)
	if err != nil {
		return nil, err
	}
	if len(days) > 0 {
		log.Printf("shipments %s<число>: заношу %s", w.prefix, FormatDays(days))
	}
	w.progress(nil, len(days))

	var written []time.Time
	var errs []error
	for _, day := range days {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		ok, err := w.writeDay(ctx, day, !day.Equal(yesterday), existing)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", day.Format("02.01"), err))
			continue
		}
		if ok {
			written = append(written, day)
			log.Printf("shipments %s<число>: внесено %s", w.prefix, day.Format("02.01"))
			w.progress(written, len(days))
		}
	}
	return written, errors.Join(errs...)
}

// emptyDays оставляет дни, листа которых нет или он пустой. Решение принимается
// только по таблице: начало всех листов читается одним запросом.
func (w ShipmentsWriter) emptyDays(days []time.Time, existing map[string]string) ([]time.Time, error) {
	var ranges []string
	for _, d := range days {
		if title, ok := w.sheetTitle(d, existing); ok {
			ranges = append(ranges, google.SheetRange(title, "A1:C3"))
		}
	}
	cells, err := w.sheets.BatchGet(w.spreadsheetID, ranges)
	if err != nil {
		return nil, fmt.Errorf("read day sheets: %w", err)
	}

	var todo []time.Time
	i := 0
	for _, d := range days {
		if _, ok := w.sheetTitle(d, existing); !ok {
			todo = append(todo, d)
			continue
		}
		if i >= len(cells) || isEmpty(cells[i]) {
			todo = append(todo, d)
		}
		i++
	}
	return todo, nil
}

// writeDay собирает блоки всех кабинетов за день и пишет их на лист дня. Если
// отгрузок нет ни у кого: за вчера лист остаётся пустым (данные могли не успеть
// появиться, следующей ночью день проверится снова), за более ранние дни пишется
// пустой блок, чтобы день больше не запрашивался.
func (w ShipmentsWriter) writeDay(ctx context.Context, day time.Time, placeholder bool, existing map[string]string) (bool, error) {
	blocks := make([]ShipmentsBlock, 0, len(w.sources))
	for _, s := range w.sources {
		b, err := s.ShipmentsBlock(ctx, day)
		if err != nil {
			return false, err
		}
		blocks = append(blocks, b)
	}

	values := shipmentsValues(blocks, placeholder)
	if len(values) == 0 {
		return false, nil
	}

	title, ok := w.sheetTitle(day, existing)
	if !ok {
		if err := w.sheets.AddSheets(w.spreadsheetID, []string{title}); err != nil {
			return false, err
		}
		existing[strings.ToLower(title)] = title
	}

	width := 0
	for _, row := range values {
		width = max(width, len(row))
	}
	cells := fmt.Sprintf("A1:%s%d", google.ColumnLetter(width-1), len(values))
	if err := w.sheets.BatchWriteUserEntered(w.spreadsheetID, []google.ValueRange{{Range: google.SheetRange(title, cells), Values: values}}); err != nil {
		return false, err
	}
	return true, nil
}

// sheetTitle — лист дня day: его настоящее название в таблице и есть ли он там вообще.
func (w ShipmentsWriter) sheetTitle(day time.Time, existing map[string]string) (string, bool) {
	title := w.prefix + strconv.Itoa(day.Day())
	if actual, ok := existing[strings.ToLower(title)]; ok {
		return actual, true
	}
	return title, false
}

func (w ShipmentsWriter) progress(written []time.Time, total int) {
	if w.Progress != nil {
		w.Progress(written, total)
	}
}

// shipmentsValues раскладывает блоки так же, как раньше их дописывал каждый кабинет:
// строка с названием кабинета, шапка, строки — и следующий блок сразу под ним.
// Кабинеты без отгрузок пропускаются. Если отгрузок нет ни у кого и нужен
// placeholder, остаётся блок первого кабинета без строк.
func shipmentsValues(blocks []ShipmentsBlock, placeholder bool) [][]interface{} {
	var out [][]interface{}
	for _, b := range blocks {
		if len(b.Rows) == 0 {
			continue
		}
		out = append(out, []interface{}{b.Cabinet}, toCells(b.Header))
		for _, row := range b.Rows {
			out = append(out, toCells(row))
		}
	}

	if len(out) == 0 && placeholder && len(blocks) > 0 {
		out = append(out, []interface{}{blocks[0].Cabinet})
		if len(blocks[0].Header) > 0 {
			out = append(out, toCells(blocks[0].Header))
		}
	}
	return out
}

func toCells(row []string) []interface{} {
	out := make([]interface{}, len(row))
	for i, v := range row {
		out[i] = v
	}
	return out
}

func isEmpty(rows [][]interface{}) bool {
	for _, row := range rows {
		for _, v := range row {
			if strings.TrimSpace(fmt.Sprint(v)) != "" {
				return false
			}
		}
	}
	return true
}
