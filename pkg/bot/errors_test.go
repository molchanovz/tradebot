package bot

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tradebot/pkg/client/apierr"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserError(t *testing.T) {
	apiErr := func(status int) error {
		return fmt.Errorf("get labels failed: %w", &apierr.Error{Market: apierr.MarketOzon, Method: "/v3/posting/fbs/list", Status: status, Body: `{"code":8}`})
	}

	t.Run("rate limit", func(t *testing.T) {
		text := userError(apiErr(http.StatusTooManyRequests))
		assert.True(t, strings.HasPrefix(text, "Озон: превышен лимит запросов в секунду"), text)
		assert.NotContains(t, text, `{"code":8}`)
	})

	t.Run("rate limit after retries", func(t *testing.T) {
		err := &apierr.Error{Market: apierr.MarketOzon, Status: http.StatusTooManyRequests, Attempts: 4}
		assert.Contains(t, userError(err), "Бот повторил запрос 3 раза, но лимит не освободился.")
		assert.NotContains(t, userError(apiErr(http.StatusTooManyRequests)), "повторил")
	})

	t.Run("bad key", func(t *testing.T) {
		assert.Contains(t, userError(apiErr(http.StatusForbidden)), "API-ключ")
		assert.Contains(t, userError(apiErr(http.StatusUnauthorized)), "API-ключ")
	})

	t.Run("server error", func(t *testing.T) {
		assert.Contains(t, userError(apiErr(http.StatusBadGateway)), "сбой на стороне маркетплейса")
	})

	t.Run("other status keeps body", func(t *testing.T) {
		assert.Equal(t, `Озон вернул ошибку (код 400): {"code":8}`, userError(apiErr(http.StatusBadRequest)))
	})

	t.Run("no orders", func(t *testing.T) {
		assert.Equal(t, "Заказов для печати нет.", userError(ozon.ErrNoRows))
	})

	t.Run("unknown error as is", func(t *testing.T) {
		assert.Equal(t, "ошибка записи PDF", userError(errors.New("ошибка записи PDF")))
	})
}

func TestFromResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"code":8,"message":"You have reached request rate limit per second"}`)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v3/posting/fbs/list")
	require.NoError(t, err)
	defer resp.Body.Close()

	e := apierr.FromResponse(apierr.MarketOzon, resp)
	assert.Equal(t, "/v3/posting/fbs/list", e.Method)
	assert.Equal(t, http.StatusTooManyRequests, e.Status)
	assert.Equal(t, `Озон /v3/posting/fbs/list: 429 Too Many Requests: {"code":8,"message":"You have reached request rate limit per second"}`, e.Error())
}

func TestTimesWord(t *testing.T) {
	for n, want := range map[int]string{1: "раз", 2: "раза", 3: "раза", 5: "раз", 11: "раз", 12: "раз", 22: "раза"} {
		assert.Equal(t, want, timesWord(n), n)
	}
}

func TestProgressText(t *testing.T) {
	assert.Equal(t, "Обработано заказов: 5 из 10", progressText(tradeplus.Progress{Current: 5, Total: 10}))
	assert.Equal(t, "⏳ Получаю список заказов из Озона…", progressText(tradeplus.Progress{Stage: "Получаю список заказов из Озона"}))
	assert.Equal(t, "⏳ Готовлю стикеры: 15 из 120\n⚠️ Озон ограничил частоту запросов, повтор 2 из 3 через 2 с…",
		progressText(tradeplus.Progress{Stage: "Готовлю стикеры", Current: 15, Total: 120, Note: "Озон ограничил частоту запросов, повтор 2 из 3 через 2 с…"}))
}
