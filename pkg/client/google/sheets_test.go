package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSpreadsheetID(t *testing.T) {
	const id = "1WOUHE2qs-c2idJN4pduWkT6PqJzX8XioI-I3ZoeGxMo"

	tests := []struct {
		name   string
		in     string
		wantID string
		wantOK bool
	}{
		{name: "edit link", in: "https://docs.google.com/spreadsheets/d/" + id + "/edit#gid=0", wantID: id, wantOK: true},
		{name: "share link", in: "https://docs.google.com/spreadsheets/d/" + id + "/edit?usp=sharing", wantID: id, wantOK: true},
		{name: "link without path", in: "https://docs.google.com/spreadsheets/d/" + id, wantID: id, wantOK: true},
		{name: "link with account index", in: "https://docs.google.com/spreadsheets/u/1/d/" + id + "/edit", wantID: id, wantOK: true},
		{name: "link inside text", in: "вот таблица: https://docs.google.com/spreadsheets/d/" + id + "/edit спасибо", wantID: id, wantOK: true},
		{name: "bare id with spaces", in: "  " + id + "\n", wantID: id, wantOK: true},
		{name: "published link is not editable", in: "https://docs.google.com/spreadsheets/d/e/2PACX-1vQ/pubhtml", wantOK: false},
		{name: "other site", in: "https://example.com/table", wantOK: false},
		{name: "plain text", in: "таблица за октябрь", wantOK: false},
		{name: "empty", in: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseSpreadsheetID(tt.in)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestSheetRange(t *testing.T) {
	assert.Equal(t, "'Заказы WB-9'!A2:B", SheetRange("Заказы WB-9", "A2:B"))
	assert.Equal(t, "'Lena''s'!A1", SheetRange("Lena's", "A1"))
}
