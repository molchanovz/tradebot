package bot

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"tradebot/pkg/client/apierr"
	"tradebot/pkg/tradeplus/ozon"
)

// userError переводит ошибку в текст для пользователя бота. Ошибки API маркетплейсов
// и таймауты получают понятное объяснение, остальные показываются как есть.
// Полный текст ошибки нужно писать в лог отдельно.
func userError(err error) string {
	if errors.Is(err, ozon.ErrNoRows) {
		return "Заказов для печати нет."
	}

	var apiErr *apierr.Error
	if errors.As(err, &apiErr) {
		return apiErrorText(apiErr)
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
		return "Маркетплейс не ответил вовремя. Попробуйте ещё раз через пару минут."
	}

	return err.Error()
}

func apiErrorText(e *apierr.Error) string {
	switch {
	case e.Status == http.StatusTooManyRequests:
		retried := ""
		if retries := e.Attempts - 1; retries > 0 {
			retried = fmt.Sprintf(" Бот повторил запрос %d %s, но лимит не освободился.", retries, timesWord(retries))
		}
		return fmt.Sprintf("%s: превышен лимит запросов в секунду для кабинета.%s Подождите минуту и попробуйте ещё раз. "+
			"Если ошибка повторяется часто, лимит может расходовать другая программа, подключённая к этому кабинету по API.", e.Market, retried)
	case e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden:
		return fmt.Sprintf("%s не принял API-ключ кабинета: ключ неверный, отозван или без нужных прав. Проверьте ключ в /settings. (код %d)",
			e.Market, e.Status)
	case e.Status == http.StatusNotFound:
		return fmt.Sprintf("%s: данные не найдены. Проверьте номер поставки или заказа. (код %d)", e.Market, e.Status)
	case e.Status >= http.StatusInternalServerError:
		return fmt.Sprintf("%s: сбой на стороне маркетплейса. Попробуйте позже. (код %d)", e.Market, e.Status)
	default:
		return fmt.Sprintf("%s вернул ошибку (код %d): %s", e.Market, e.Status, e.Body)
	}
}

// timesWord — «раз» или «раза» после числа n.
func timesWord(n int) string {
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 12 || n%100 > 14) {
		return "раза"
	}
	return "раз"
}
