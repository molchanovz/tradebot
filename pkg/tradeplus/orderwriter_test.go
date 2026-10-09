package tradeplus

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"tradebot/pkg/client/google"
	"tradebot/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrdersDays(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{name: "middle of month: from the 1st", now: time.Date(2026, 10, 9, 8, 0, 0, 0, MSK), want: "01.10, 02.10, 03.10, 04.10, 05.10, 06.10, 07.10, 08.10"},
		{name: "early in month: previous month is skipped", now: time.Date(2026, 10, 3, 8, 0, 0, 0, MSK), want: "01.10, 02.10"},
		{name: "second day: only yesterday", now: time.Date(2026, 10, 2, 8, 0, 0, 0, MSK), want: "01.10"},
		{name: "first day: yesterday from previous month", now: time.Date(2026, 10, 1, 8, 0, 0, 0, MSK), want: "30.09"},
		{name: "last day of month", now: time.Date(2026, 2, 28, 8, 0, 0, 0, MSK), want: "01.02, 02.02, 03.02, 04.02, 05.02, 06.02, 07.02, 08.02, 09.02, 10.02, 11.02, 12.02, 13.02, 14.02, 15.02, 16.02, 17.02, 18.02, 19.02, 20.02, 21.02, 22.02, 23.02, 24.02, 25.02, 26.02, 27.02"},
		{name: "days are counted in MSK", now: time.Date(2026, 10, 8, 22, 30, 0, 0, time.UTC), want: "01.10, 02.10, 03.10, 04.10, 05.10, 06.10, 07.10, 08.10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatDays(OrdersDays(tt.now)))
		})
	}
}

func TestOrdersReport_valueRanges(t *testing.T) {
	day := time.Date(2026, 10, 8, 0, 0, 0, 0, MSK)

	t.Run("one section", func(t *testing.T) {
		r := OrdersReport{Day: day, Sections: []OrdersSection{{Blocks: []OrdersBlock{
			{Title: "Заказы FBO", Counts: map[string]int{"b": 2, "a": 1}},
			{Title: "Заказы FBS", Counts: map[string]int{}},
			{Title: "Возвраты", Counts: map[string]int{"c": 5}},
		}}}}

		assert.Equal(t, []google.ValueRange{
			{Range: "'Заказы WB-8'!A1", Values: [][]interface{}{{"Отчет за 08.10.2026"}}},
			{Range: "'Заказы WB-8'!A2:B4", Values: [][]interface{}{{"Заказы FBO"}, {"a", 1}, {"b", 2}}},
			{Range: "'Заказы WB-8'!D2:E2", Values: [][]interface{}{{"Заказы FBS"}}},
			{Range: "'Заказы WB-8'!G2:H3", Values: [][]interface{}{{"Возвраты"}, {"c", 5}}},
		}, r.valueRanges("Заказы WB-8"))

		assert.Equal(t, []string{"'Заказы WB-8'!A1", "'Заказы WB-8'!A2:B", "'Заказы WB-8'!D2:E", "'Заказы WB-8'!G2:H"}, r.clearRanges("Заказы WB-8"))
	})

	// Вторая секция (кабинет Ozon) начинается через пустую строку после самого
	// длинного блока первой — так раньше считал WriteOzon: maxValuesCount+3.
	t.Run("sections one under another", func(t *testing.T) {
		r := OrdersReport{Day: day, Sections: []OrdersSection{
			{Blocks: []OrdersBlock{
				{Title: "Заказы FBS", Counts: map[string]int{"x": 1, "y": 2}},
				{Title: "Заказы FBO", Counts: map[string]int{}},
			}},
			{Blocks: []OrdersBlock{
				{Title: "Заказы FBS", Counts: map[string]int{"z": 3}},
				{Title: "Заказы FBO", Counts: map[string]int{"w": 4}},
			}},
		}}

		assert.Equal(t, []google.ValueRange{
			{Range: "'Заказы OZON-8'!A1", Values: [][]interface{}{{"Отчет за 08.10.2026"}}},
			{Range: "'Заказы OZON-8'!A2:B4", Values: [][]interface{}{{"Заказы FBS"}, {"x", 1}, {"y", 2}}},
			{Range: "'Заказы OZON-8'!D2:E2", Values: [][]interface{}{{"Заказы FBO"}}},
			{Range: "'Заказы OZON-8'!A6", Values: [][]interface{}{{"Отчет за 08.10.2026"}}},
			{Range: "'Заказы OZON-8'!A7:B8", Values: [][]interface{}{{"Заказы FBS"}, {"z", 3}}},
			{Range: "'Заказы OZON-8'!D7:E8", Values: [][]interface{}{{"Заказы FBO"}, {"w", 4}}},
		}, r.valueRanges("Заказы OZON-8"))
	})
}

func TestOrdersWriter_Sync(t *testing.T) {
	now := time.Date(2026, 10, 9, 8, 0, 0, 0, MSK) // проверяются 01.10–08.10

	newSheets := func() *fakeSheets {
		return &fakeSheets{a1: map[string]string{
			"Заказы WB-1": "Отчет за 01.10.2026", // заполнен
			"Заказы WB-2": "Отчет за 02.10.2026",
			"заказы wb-3": "Отчет за 03.09.2026", // прошлый месяц, регистр в названии другой
			"Заказы WB-4": "",                    // пустой
			// листа «Заказы WB-5» нет
			"Заказы WB-6": "Отчет за 06.10.2026",
			"Заказы WB-7": "Отчет за 07.10.2026",
			"Заказы WB-8": "Отчет за 08.09.2026", // вчера, ещё не заполнен
		}}
	}

	t.Run("fills missed days and yesterday", func(t *testing.T) {
		sheets := newSheets()
		source := &fakeSource{fail: map[string]bool{"04.10": true}}
		w := OrdersWriter{spreadsheetID: "book", prefix: "Заказы WB-", source: source, sheets: sheets}

		written, err := w.Sync(t.Context(), now)
		require.ErrorContains(t, err, "04.10")

		assert.Equal(t, "03.10, 05.10, 08.10", FormatDays(written))
		assert.Equal(t, []string{"03.10", "04.10", "05.10", "08.10"}, source.calls)
		assert.Equal(t, []string{
			"'Заказы WB-1'!A1", "'Заказы WB-2'!A1", "'заказы wb-3'!A1", "'Заказы WB-4'!A1",
			"'Заказы WB-6'!A1", "'Заказы WB-7'!A1", "'Заказы WB-8'!A1",
		}, sheets.gets)
		assert.Equal(t, []string{"Заказы WB-5"}, sheets.added)
		assert.Equal(t, []string{
			"'заказы wb-3'!A1", "'заказы wb-3'!A2:B",
			"'Заказы WB-5'!A1", "'Заказы WB-5'!A2:B",
			"'Заказы WB-8'!A1", "'Заказы WB-8'!A2:B",
		}, sheets.cleared)
		assert.Equal(t, map[string]string{
			"'заказы wb-3'!A1": "Отчет за 03.10.2026",
			"'Заказы WB-5'!A1": "Отчет за 05.10.2026",
			"'Заказы WB-8'!A1": "Отчет за 08.10.2026",
		}, sheets.writtenTitles())
	})

	t.Run("nothing to write when the table is complete", func(t *testing.T) {
		sheets := newSheets()
		sheets.a1["заказы wb-3"] = "Отчет за 03.10.2026"
		sheets.a1["Заказы WB-4"] = "Отчет за 04.10.2026"
		sheets.a1["Заказы WB-5"] = "Отчет за 05.10.2026"
		sheets.a1["Заказы WB-8"] = "Отчет за 08.10.2026"
		source := &fakeSource{}
		w := OrdersWriter{spreadsheetID: "book", prefix: "Заказы WB-", source: source, sheets: sheets}

		written, err := w.Sync(t.Context(), now)
		require.NoError(t, err)

		assert.Empty(t, written)
		assert.Empty(t, source.calls)
		assert.Empty(t, sheets.cleared)
		assert.Empty(t, sheets.written)
	})

	t.Run("stops between days when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		sheets := newSheets()
		source := &fakeSource{}
		w := OrdersWriter{spreadsheetID: "book", prefix: "Заказы WB-", dayInterval: time.Hour, source: source, sheets: sheets}

		written, err := w.Sync(ctx, now)
		require.ErrorIs(t, err, context.Canceled)

		assert.Equal(t, "03.10", FormatDays(written))
		assert.Equal(t, []string{"03.10"}, source.calls)
		assert.Equal(t, map[string]string{"'заказы wb-3'!A1": "Отчет за 03.10.2026"}, sheets.writtenTitles())
	})

	t.Run("no spreadsheet", func(t *testing.T) {
		w := OrdersWriter{prefix: "Заказы WB-", source: &fakeSource{}, sheets: newSheets()}

		_, err := w.Sync(t.Context(), now)
		require.Error(t, err)
	})
}

func TestCabinets_OrdersSpreadsheetIDs(t *testing.T) {
	const (
		id1 = "1WOUHE2qs-c2idJN4pduWkT6PqJzX8XioI-I3ZoeGxMo"
		id2 = "1Rljs-bxCQCP0DnDfqRGw0SKm1OjRXi2EdJrW5j3M5ts"
	)
	cabinet := func(link string, statusID int) Cabinet {
		c := Cabinet{}
		c.StatusID = statusID
		if link != "" {
			c.SheetLink = &link
		}
		return c
	}

	cc := Cabinets{
		cabinet("", db.StatusEnabled),
		cabinet("https://docs.google.com/spreadsheets/d/"+id1+"/edit", db.StatusEnabled),
		cabinet(id1, db.StatusEnabled),
		cabinet(id2, db.StatusEnabled),
	}
	assert.Equal(t, []string{id1, id2}, cc.OrdersSpreadsheetIDs())
	assert.Equal(t, id1, cc.OrdersSpreadsheetID())
	assert.Empty(t, Cabinets{cabinet("", db.StatusEnabled)}.OrdersSpreadsheetID())

	// выключенный кабинет в отчёт не попадает, остальные идут по возрастанию ID
	withID := func(c Cabinet, id int) Cabinet {
		c.ID = id
		return c
	}
	mixed := Cabinets{
		withID(cabinet(id1, db.StatusEnabled), 6),
		withID(cabinet(id2, db.StatusDisabled), 4),
		withID(cabinet(id1, db.StatusEnabled), 5),
	}
	got := mixed.OrdersCabinets()
	require.Len(t, got, 2)
	assert.Equal(t, []int{5, 6}, []int{got[0].ID, got[1].ID})
	assert.Equal(t, []string{id1}, got.OrdersSpreadsheetIDs())
}

// fakeSheets — таблица в памяти: лист → значение A1 (нет ключа — нет листа).
type fakeSheets struct {
	a1      map[string]string
	gets    []string
	added   []string
	cleared []string
	written []google.ValueRange
}

func (f *fakeSheets) Spreadsheet(string) (string, []google.SheetStat, error) {
	out := make([]google.SheetStat, 0, len(f.a1))
	for title := range f.a1 {
		out = append(out, google.SheetStat{Title: title})
	}
	return "Заказы", out, nil
}

func (f *fakeSheets) BatchGet(_ string, ranges []string) ([][][]interface{}, error) {
	f.gets = append(f.gets, ranges...)
	out := make([][][]interface{}, len(ranges))
	for i, r := range ranges {
		title := strings.TrimSuffix(strings.TrimPrefix(r, "'"), "'!A1")
		if v := f.a1[title]; v != "" {
			out[i] = [][]interface{}{{v}}
		}
	}
	return out, nil
}

func (f *fakeSheets) AddSheets(_ string, titles []string) error {
	f.added = append(f.added, titles...)
	return nil
}

func (f *fakeSheets) BatchClear(_ string, ranges []string) error {
	f.cleared = append(f.cleared, ranges...)
	return nil
}

func (f *fakeSheets) BatchWrite(_ string, data []google.ValueRange) error {
	f.written = append(f.written, data...)
	return nil
}

// writtenTitles — что записано в A1 листов.
func (f *fakeSheets) writtenTitles() map[string]string {
	out := map[string]string{}
	for _, vr := range f.written {
		if strings.HasSuffix(vr.Range, "!A1") {
			out[vr.Range] = firstCell(vr.Values)
		}
	}
	return out
}

// fakeSource отдаёт отчёт с одним блоком; для дней из fail — ошибку.
type fakeSource struct {
	fail  map[string]bool
	calls []string
}

func (s *fakeSource) OrdersReport(_ context.Context, day time.Time) (OrdersReport, error) {
	d := day.Format("02.01")
	s.calls = append(s.calls, d)
	if s.fail[d] {
		return OrdersReport{}, errors.New("api is down")
	}
	return OrdersReport{Day: day, Sections: []OrdersSection{{Blocks: []OrdersBlock{
		{Title: "Заказы FBO", Counts: map[string]int{"a": 1}},
	}}}}, nil
}
