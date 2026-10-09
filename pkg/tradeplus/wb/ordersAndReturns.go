package wb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"tradebot/pkg/client/wb"
	"tradebot/pkg/tradeplus"
)

const (
	// OrdersSheetPrefix — листы ежедневного отчёта WB: «Заказы WB-<число>».
	OrdersSheetPrefix = "Заказы WB-"
	// ordersDayInterval — пауза между днями: статистика WB отдаёт не больше
	// одного запроса в минуту на метод, а за каждый день нужны заказы и продажи.
	ordersDayInterval = time.Minute + 5*time.Second
)

// OrdersManager собирает ежедневный отчёт WB: заказы FBO, FBS и возвраты.
type OrdersManager struct {
	client wb.Client
}

func NewOrdersManager(token string) OrdersManager {
	return OrdersManager{client: wb.NewClient(token)}
}

// NewOrdersWriter — запись отчётов WB в таблицу заказов. Отчёт, как и раньше,
// строится по первому кабинету WB.
func NewOrdersWriter(cabinets tradeplus.Cabinets) (tradeplus.OrdersWriter, error) {
	if len(cabinets) == 0 {
		return tradeplus.OrdersWriter{}, errors.New("нет кабинетов WB")
	}
	return tradeplus.NewOrdersWriter(cabinets.OrdersSpreadsheetID(), OrdersSheetPrefix, ordersDayInterval, NewOrdersManager(cabinets[0].Key)), nil
}

// OrdersReport — заказы и возвраты WB за сутки day.
func (m OrdersManager) OrdersReport(ctx context.Context, day time.Time) (tradeplus.OrdersReport, error) {
	postingsWithCountFBO, postingsWithCountFBS, err := m.getPostingsMap(ctx, day)
	if err != nil {
		return tradeplus.OrdersReport{}, err
	}

	returnsWithCount, err := m.getReturnsMap(ctx, day)
	if err != nil {
		return tradeplus.OrdersReport{}, err
	}

	return tradeplus.OrdersReport{Day: day, Sections: []tradeplus.OrdersSection{{Blocks: []tradeplus.OrdersBlock{
		{Title: "Заказы FBO", Counts: postingsWithCountFBO},
		{Title: "Заказы FBS", Counts: postingsWithCountFBS},
		{Title: "Возвраты", Counts: returnsWithCount},
	}}}}, nil
}

func (m OrdersManager) getPostingsMap(ctx context.Context, day time.Time) (map[string]int, map[string]int, error) {
	postingsWithCountFBO := make(map[string]int)
	postingsWithCountFBS := make(map[string]int)

	postingsList, err := m.client.OrdersOnDate(ctx, day)
	if err != nil {
		return nil, nil, fmt.Errorf("wb orders failed: %w", err)
	}

	for _, posting := range postingsList {
		if !posting.IsCancel {
			switch posting.WarehouseType {
			case "Склад WB":
				postingsWithCountFBO[posting.SupplierArticle]++
			case "Склад продавца":
				postingsWithCountFBS[posting.SupplierArticle]++
			default:
				if strings.Contains(posting.WarehouseName, "мп") {
					postingsWithCountFBS[posting.SupplierArticle]++
				} else {
					postingsWithCountFBO[posting.SupplierArticle]++
				}
			}
		}
	}

	return postingsWithCountFBO, postingsWithCountFBS, nil
}

func (m OrdersManager) getReturnsMap(ctx context.Context, day time.Time) (map[string]int, error) {
	returnsWithCount := make(map[string]int)

	returnsList, err := m.client.SalesOnDate(ctx, day)
	if err != nil {
		return nil, fmt.Errorf("wb sales failed: %w", err)
	}

	for _, someReturn := range returnsList {
		if strings.HasPrefix(someReturn.SaleID, "R") {
			returnsWithCount[someReturn.SupplierArticle]++
		}
	}

	return returnsWithCount, nil
}
