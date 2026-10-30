package tradeplus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"tradebot/pkg/client/google"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentsValues(t *testing.T) {
	header := []string{"№", "Артикул продавца"}
	trade := ShipmentsBlock{Cabinet: "Трейд-Плюс", Header: header, Rows: [][]string{{"1", "A"}, {"2", "B"}}}
	chem := ShipmentsBlock{Cabinet: "Химия Трейд-Плюс", Header: header, Rows: [][]string{{"3", "C"}}}
	empty := ShipmentsBlock{Cabinet: "Пустой", Header: header}

	// Блоки идут подряд, как раньше их дописывали кабинеты: шапка каждого блока — во второй его строке.
	assert.Equal(t, [][]interface{}{
		{"Трейд-Плюс"}, {"№", "Артикул продавца"}, {"1", "A"}, {"2", "B"},
		{"Химия Трейд-Плюс"}, {"№", "Артикул продавца"}, {"3", "C"},
	}, shipmentsValues([]ShipmentsBlock{trade, empty, chem}, true))

	assert.Equal(t, [][]interface{}{{"Пустой"}, {"№", "Артикул продавца"}}, shipmentsValues([]ShipmentsBlock{empty, empty}, true))
	assert.Empty(t, shipmentsValues([]ShipmentsBlock{empty}, false))
	assert.Equal(t, [][]interface{}{{"Пустой"}}, shipmentsValues([]ShipmentsBlock{{Cabinet: "Пустой"}}, true))
}

func TestShipmentsWriter_Sync(t *testing.T) {
	now := time.Date(2026, 10, 9, 8, 0, 0, 0, MSK) // проверяются 01.10–08.10, вчера — 08.10

	sheets := &fakeShipmentSheets{filled: map[string]bool{
		"Отправлено на WB FBS-1": true,  // заполнен — не трогаем
		"Отправлено на WB FBS-2": false, // пустой
		"отправлено на wb fbs-3": false, // пустой, регистр в названии другой
		"Отправлено на WB FBS-4": true,
		// листа «Отправлено на WB FBS-5» нет
		"Отправлено на WB FBS-6": false, // отгрузок не было — пишем пустой блок
		"Отправлено на WB FBS-7": false, // источник упал — не пишем
		"Отправлено на WB FBS-8": false, // вчера без отгрузок — оставляем до следующей ночи
	}}
	trade := &fakeShipmentSource{cabinet: "Трейд-Плюс", rows: map[string][][]string{
		"02.10": {{"1", "A"}}, "03.10": {{"2", "B"}}, "05.10": {{"3", "C"}},
	}, fail: map[string]bool{"07.10": true}}
	chem := &fakeShipmentSource{cabinet: "Химия", rows: map[string][][]string{"03.10": {{"4", "D"}}}}

	var progress []string
	w := ShipmentsWriter{spreadsheetID: "warehouse", prefix: "Отправлено на WB FBS-", sources: []ShipmentsSource{trade, chem}, sheets: sheets}
	w.Progress = func(written []time.Time, total int) {
		progress = append(progress, fmt.Sprintf("%d/%d", len(written), total))
	}

	written, err := w.Sync(t.Context(), now)
	require.ErrorContains(t, err, "07.10")

	assert.Equal(t, "02.10, 03.10, 05.10, 06.10", FormatDays(written))
	assert.Equal(t, []string{"Отправлено на WB FBS-5"}, sheets.added)
	assert.Equal(t, []string{"0/6", "1/6", "2/6", "3/6", "4/6"}, progress)
	assert.Equal(t, []google.ValueRange{
		{Range: "'Отправлено на WB FBS-2'!A1:B3", Values: [][]interface{}{{"Трейд-Плюс"}, {"№", "Артикул продавца"}, {"1", "A"}}},
		{Range: "'отправлено на wb fbs-3'!A1:B6", Values: [][]interface{}{
			{"Трейд-Плюс"}, {"№", "Артикул продавца"}, {"2", "B"},
			{"Химия"}, {"№", "Артикул продавца"}, {"4", "D"},
		}},
		{Range: "'Отправлено на WB FBS-5'!A1:B3", Values: [][]interface{}{{"Трейд-Плюс"}, {"№", "Артикул продавца"}, {"3", "C"}}},
		{Range: "'Отправлено на WB FBS-6'!A1:B2", Values: [][]interface{}{{"Трейд-Плюс"}, {"№", "Артикул продавца"}}},
	}, sheets.written)
}

func TestNewShipmentsWriters(t *testing.T) {
	const (
		id1 = "1WOUHE2qs-c2idJN4pduWkT6PqJzX8XioI-I3ZoeGxMo"
		id2 = "1Rljs-bxCQCP0DnDfqRGw0SKm1OjRXi2EdJrW5j3M5ts"
	)
	cabinet := func(id int, sheet string) Cabinet {
		c := Cabinet{}
		c.ID = id
		c.Name = fmt.Sprint("кабинет ", id)
		c.Settings.ShipmentsSheetID = sheet
		return c
	}
	source := func(_ context.Context, c Cabinet) (ShipmentsSource, error) {
		return &fakeShipmentSource{cabinet: c.Name}, nil
	}

	writers, err := NewShipmentsWriters(t.Context(), Cabinets{cabinet(1, id1), cabinet(2, ""), cabinet(3, id1), cabinet(4, id2)}, "Отправлено на WB FBS-", source)
	require.NoError(t, err)
	require.Len(t, writers, 2)
	assert.Equal(t, id1, writers[0].spreadsheetID)
	assert.Len(t, writers[0].sources, 2)
	assert.Equal(t, id2, writers[1].spreadsheetID)
	assert.Len(t, writers[1].sources, 1)

	_, err = NewShipmentsWriters(t.Context(), Cabinets{cabinet(1, id1)}, "", func(context.Context, Cabinet) (ShipmentsSource, error) {
		return nil, errors.New("нет ключа")
	})
	require.ErrorContains(t, err, "нет ключа")
}

// fakeShipmentSheets — складская таблица в памяти: лист → есть ли на нём данные (нет ключа — нет листа).
type fakeShipmentSheets struct {
	filled  map[string]bool
	added   []string
	written []google.ValueRange
}

func (f *fakeShipmentSheets) Spreadsheet(string) (string, []google.SheetStat, error) {
	out := make([]google.SheetStat, 0, len(f.filled))
	for title := range f.filled {
		out = append(out, google.SheetStat{Title: title})
	}
	return "Складская таблица", out, nil
}

func (f *fakeShipmentSheets) BatchGet(_ string, ranges []string) ([][][]interface{}, error) {
	out := make([][][]interface{}, len(ranges))
	for i, r := range ranges {
		title := strings.TrimSuffix(strings.TrimPrefix(r, "'"), "'!A1:C3")
		if f.filled[title] {
			out[i] = [][]interface{}{{"Трейд-Плюс"}}
		}
	}
	return out, nil
}

func (f *fakeShipmentSheets) AddSheets(_ string, titles []string) error {
	f.added = append(f.added, titles...)
	return nil
}

func (f *fakeShipmentSheets) BatchWriteUserEntered(_ string, data []google.ValueRange) error {
	f.written = append(f.written, data...)
	return nil
}

// fakeShipmentSource — отгрузки кабинета по дням «02.01»; для дней из fail — ошибка.
type fakeShipmentSource struct {
	cabinet string
	rows    map[string][][]string
	fail    map[string]bool
}

func (s *fakeShipmentSource) ShipmentsBlock(_ context.Context, day time.Time) (ShipmentsBlock, error) {
	d := day.Format("02.01")
	if s.fail[d] {
		return ShipmentsBlock{}, errors.New("api is down")
	}
	return ShipmentsBlock{Cabinet: s.cabinet, Header: []string{"№", "Артикул продавца"}, Rows: s.rows[d]}, nil
}
