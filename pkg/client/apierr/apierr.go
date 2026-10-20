// Package apierr описывает неуспешный ответ API маркетплейса: по HTTP-статусу бот
// подбирает понятный пользователю текст, а в лог уходит полное техническое описание.
package apierr

import (
	"fmt"
	"io"
	"net/http"
)

const (
	MarketOzon   = "Озон"
	MarketWB     = "WB"
	MarketYandex = "Яндекс Маркет"
)

// maxBody — сколько байт тела ответа сохраняем в ошибке.
const maxBody = 2048

type Error struct {
	Market string // название маркетплейса для пользователя, MarketOzon и т.п.
	Method string // путь запроса, например /v3/posting/fbs/list
	Status int
	Body   string
	// Attempts — сколько раз отправили запрос, считая повторы; 0 и 1 — без повторов.
	Attempts int
}

// FromResponse читает тело неуспешного ответа resp и собирает из него Error.
func FromResponse(market string, resp *http.Response) *Error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	return &Error{
		Market: market,
		Method: resp.Request.URL.Path,
		Status: resp.StatusCode,
		Body:   string(body),
	}
}

func (e *Error) Error() string {
	msg := fmt.Sprintf("%s %s: %d %s: %s", e.Market, e.Method, e.Status, http.StatusText(e.Status), e.Body)
	if e.Attempts > 1 {
		msg += fmt.Sprintf(" (попыток: %d)", e.Attempts)
	}
	return msg
}
