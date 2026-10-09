package wb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (c Client) GetOrderIDsFbs(supplyID string) (OrderIDs, error) {
	var orders OrderIDs
	jsonString, err := c.getOrdersBySupplyID(supplyID)
	if err != nil || jsonString == "" {
		return orders, err
	}

	err = json.Unmarshal([]byte(jsonString), &orders)
	if err != nil {
		return orders, err
	}

	//sortOrdersByArticle(orders.Orders)
	return orders, nil
}
func (c Client) GetCards(nmID *int, updatedAt *time.Time, limit *int) (*CardList, error) {
	var cards CardList
	jsonString, err := c.getCards(nmID, updatedAt, limit)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &cards)
	if err != nil {
		return nil, err
	}

	return &cards, nil
}

func (c Client) GetReturns(dateFrom, dateTo string) (*ReturnList, error) {
	var returns ReturnList
	jsonString, err := c.getReturns(dateFrom, dateTo)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &returns)
	if err != nil {
		return nil, err
	}

	return &returns, nil
}

func (c Client) GetStickersFbs(orderID int) (StickerWB, error) {
	var stickers StickerWB
	jsonString, err := c.getCodesByOrderID(orderID)
	if err != nil || jsonString == "" {
		return stickers, err
	}

	err = json.Unmarshal([]byte(jsonString), &stickers)
	return stickers, err
}

func (c Client) GetAllOrders(daysAgo, flag int) (OrdersListALL, error) {
	var posting OrdersListALL
	jsonString, err := c.apiOrdersALL(daysAgo, flag)
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &posting)
	return posting, err
}

// OrdersOnDate возвращает заказы из статистики WB с датой date.
func (c Client) OrdersOnDate(ctx context.Context, date time.Time) (OrdersListALL, error) {
	var orders OrdersListALL
	err := c.statisticsJSON(ctx, "/api/v1/supplier/orders", date, &orders)
	return orders, err
}

// SalesOnDate возвращает продажи и возвраты из статистики WB с датой date.
func (c Client) SalesOnDate(ctx context.Context, date time.Time) (SalesReturns, error) {
	var sales SalesReturns
	err := c.statisticsJSON(ctx, "/api/v1/supplier/sales", date, &sales)
	return sales, err
}

func (c Client) statisticsJSON(ctx context.Context, path string, date time.Time, out any) error {
	body, err := c.statisticsOnDate(ctx, path, date)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func (c Client) GetPostingStatus(postingID int) (string, error) {
	var postingStatuses OrdersWithStatuses
	jsonString, err := c.ordersFBSStatus(postingID)
	if err != nil || jsonString == "" {
		return "", err
	}

	err = json.Unmarshal([]byte(jsonString), &postingStatuses)
	if err != nil {
		return "", err
	}

	return postingStatuses.Orders[0].WbStatus, nil
}

func (c Client) GetStockFbo() ([]Stock, error) {
	var stocks []Stock
	jsonString, err := c.stocksFbo()
	if err != nil || jsonString == "" {
		return nil, err
	}

	err = json.Unmarshal([]byte(jsonString), &stocks)
	return stocks, nil
}
