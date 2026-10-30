package wb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"tradebot/pkg/client/apierr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetReturns(t *testing.T) {
	dateFrom := time.Date(2026, 10, 1, 15, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC)

	t.Run("collects all pages", func(t *testing.T) {
		const total = returnsPageSize + 2
		var offsets []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, returnsPath, r.URL.Path)
			assert.Equal(t, "token", r.Header.Get("Authorization"))
			assert.Equal(t, "2026-10-01", q.Get("dateFrom"))
			assert.Equal(t, "2026-10-30", q.Get("dateTo"))
			assert.Equal(t, ReturnsActive, q.Get("status"))
			assert.Equal(t, strconv.Itoa(returnsPageSize), q.Get("limit"))
			offsets = append(offsets, q.Get("offset"))

			offset, _ := strconv.Atoi(q.Get("offset"))
			report := make([]map[string]any, 0, returnsPageSize)
			for i := offset; i < min(offset+returnsPageSize, total); i++ {
				report = append(report, map[string]any{
					"nmId":             i,
					"sku":              "sku" + strconv.Itoa(i),
					"returnStatus":     "Готов к выдаче",
					"dstOfficeAddress": "Москва, Тверская 1",
				})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"count": total, "report": report})
		}))
		defer srv.Close()

		got, err := NewClient("token").WithAnalyticsURL(srv.URL).GetReturns(t.Context(), dateFrom, dateTo, ReturnsActive)
		require.NoError(t, err)
		assert.Equal(t, []string{"0", strconv.Itoa(returnsPageSize)}, offsets)
		assert.Equal(t, total, got.Count)
		require.Len(t, got.Report, total)
		last := got.Report[total-1]
		assert.Equal(t, total-1, last.NmId)
		assert.Equal(t, "sku"+strconv.Itoa(total-1), last.Sku)
		assert.Equal(t, "Готов к выдаче", last.ReturnStatus)
		assert.Equal(t, "Москва, Тверская 1", last.DstOfficeAddress)
	})

	t.Run("204 means no returns", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		got, err := NewClient("token").WithAnalyticsURL(srv.URL).GetReturns(t.Context(), dateFrom, dateTo, ReturnsActive)
		require.NoError(t, err)
		assert.Empty(t, got.Report)
	})

	t.Run("api error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"title":"not found"}`, http.StatusNotFound)
		}))
		defer srv.Close()

		_, err := NewClient("token").WithAnalyticsURL(srv.URL).GetReturns(t.Context(), dateFrom, dateTo, ReturnsActive)
		var apiErr *apierr.Error
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.Status)
		assert.Equal(t, returnsPath, apiErr.Method)
	})
}
