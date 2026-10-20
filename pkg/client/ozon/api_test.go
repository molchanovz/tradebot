package ozon

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tradebot/pkg/client/apierr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRequestRateLimit(t *testing.T) {
	var delays []time.Duration
	sleep = func(d time.Duration) { delays = append(delays, d) }
	t.Cleanup(func() { sleep = time.Sleep })

	// server отвечает 429 на первые limited запросов, потом 200.
	server := func(t *testing.T, limited int) (*httptest.Server, *int) {
		calls := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			body, _ := io.ReadAll(r.Body)
			assert.JSONEq(t, `{"posting_number":["1"]}`, string(body), "тело запроса уходит и при повторе")
			if calls <= limited {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = io.WriteString(w, `{"code":8,"message":"You have reached request rate limit per second"}`)
				return
			}
			_, _ = io.WriteString(w, `{"result":"ok"}`)
		}))
		t.Cleanup(srv.Close)
		return srv, &calls
	}

	t.Run("retries until success", func(t *testing.T) {
		delays = nil
		srv, calls := server(t, 2)

		resp, err := NewClient("id", "key").post(srv.URL, nil, nil, []byte(`{"posting_number":["1"]}`))
		require.NoError(t, err)
		assert.JSONEq(t, `{"result":"ok"}`, resp)
		assert.Equal(t, 3, *calls)
		require.Len(t, delays, 2)
		assert.GreaterOrEqual(t, delays[0], time.Second)
		assert.GreaterOrEqual(t, delays[1], 2*time.Second)
	})

	t.Run("gives up after retries", func(t *testing.T) {
		delays = nil
		srv, calls := server(t, 100)

		_, err := NewClient("id", "key").post(srv.URL, nil, nil, []byte(`{"posting_number":["1"]}`))
		var apiErr *apierr.Error
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusTooManyRequests, apiErr.Status)
		assert.Equal(t, RateLimitRetries+1, *calls)
		assert.Len(t, delays, RateLimitRetries)
	})
}

func TestRateLimitDelay(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "5")
	d := RateLimitDelay(0, h)
	assert.GreaterOrEqual(t, d, 5*time.Second)
	assert.Less(t, d, 5*time.Second+rateLimitJitter)

	d = RateLimitDelay(2, http.Header{})
	assert.GreaterOrEqual(t, d, 3*time.Second)
	assert.Less(t, d, 3*time.Second+rateLimitJitter)
}

func TestClientRateLimitHook(t *testing.T) {
	sleep = func(time.Duration) {}
	t.Cleanup(func() { sleep = time.Sleep })

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	var retries []int
	c := NewClient("id", "key").WithRateLimitHook(func(retry int, delay time.Duration) {
		retries = append(retries, retry)
		assert.Positive(t, delay)
	})
	_, err := c.post(srv.URL, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, retries)
}
