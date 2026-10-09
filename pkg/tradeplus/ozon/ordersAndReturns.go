package ozon

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tradebot/pkg/client/ozon"
	"tradebot/pkg/tradeplus"
)

// OrdersSheetPrefix — листы ежедневного отчёта Ozon: «Заказы OZON-<число>».
const OrdersSheetPrefix = "Заказы OZON-"

// OrdersManager собирает секцию ежедневного отчёта одного кабинета Ozon:
// заказы FBS, FBO и возвраты.
type OrdersManager struct {
	clientID, token string
}

func NewOrdersManager(clientID, token string) OrdersManager {
	return OrdersManager{clientID: clientID, token: token}
}

// NewOrdersWriter — запись отчётов Ozon в таблицу заказов: секция на каждый
// кабинет, секции друг под другом на одном листе.
func NewOrdersWriter(cabinets tradeplus.Cabinets) (tradeplus.OrdersWriter, error) {
	if len(cabinets) == 0 {
		return tradeplus.OrdersWriter{}, errors.New("нет кабинетов Ozon")
	}

	managers := make(cabinetsOrders, 0, len(cabinets))
	for _, cabinet := range cabinets {
		managers = append(managers, NewService(cabinet).GetOrdersAndReturnsManager())
	}
	return tradeplus.NewOrdersWriter(cabinets.OrdersSpreadsheetID(), OrdersSheetPrefix, 0, managers), nil
}

// section — блоки кабинета за сутки day.
func (m OrdersManager) section(day time.Time) (tradeplus.OrdersSection, error) {
	// сутки по Москве
	since := day.AddDate(0, 0, -1).Format("2006-01-02") + "T21:00:00.000Z"
	to := day.Format("2006-01-02") + "T21:00:00.000Z"

	postingsWithCountFBS, err := m.getPostingsMapFBS(since, to)
	if err != nil {
		return tradeplus.OrdersSection{}, fmt.Errorf("fbs: %w", err)
	}

	postingsWithCountFBO, err := m.getPostingsMapFBO(since, to)
	if err != nil {
		return tradeplus.OrdersSection{}, fmt.Errorf("fbo: %w", err)
	}

	returnsWithCount, err := m.GetReturnsMap(m.clientID, m.token, since, to)
	if err != nil {
		return tradeplus.OrdersSection{}, fmt.Errorf("returns: %w", err)
	}

	return tradeplus.OrdersSection{Blocks: []tradeplus.OrdersBlock{
		{Title: "Заказы FBS", Counts: postingsWithCountFBS},
		{Title: "Заказы FBO", Counts: postingsWithCountFBO},
		{Title: "Возвраты", Counts: returnsWithCount},
	}}, nil
}

func (m OrdersManager) getPostingsMapFBS(since, to string) (map[string]int, error) {
	postingsWithCountFBS := make(map[string]int)

	client := ozon.NewClient(m.clientID, m.token)
	for offset := 0; ; offset += ozon.PostingsListLimit {
		postingsListFbs, err := client.PostingsListFbs(since, to, offset, "")
		if err != nil {
			return nil, err
		}

		for _, posting := range postingsListFbs.Result.PostingsFBS {
			if posting.Status != "cancelled" {
				for _, product := range posting.Products {
					postingsWithCountFBS[product.OfferID] += product.Quantity
				}
			}
		}

		if !postingsListFbs.Result.HasNext {
			return postingsWithCountFBS, nil
		}
	}
}

func (m OrdersManager) getPostingsMapFBO(since, to string) (map[string]int, error) {
	postingsWithCountFBO := make(map[string]int)

	client := ozon.NewClient(m.clientID, m.token)
	for offset := 0; ; offset += ozon.PostingsListLimit {
		postingsListFbo, err := client.PostingsListFbo(since, to, offset)
		if err != nil {
			return nil, err
		}

		for _, posting := range postingsListFbo.Result {
			if posting.Status != "cancelled" {
				for _, product := range posting.Products {
					postingsWithCountFBO[product.OfferID] += product.Quantity
				}
			}
		}

		if len(postingsListFbo.Result) < ozon.PostingsListLimit {
			return postingsWithCountFBO, nil
		}
	}
}

func (m OrdersManager) GetReturnsMap(clientID, token, since, to string) (map[string]int, error) {
	var lastID int
	hasNext := true
	returnsWithCount := make(map[string]int)
	/*
		Лимит у запроса 1000, но нам нужны все возвраты,
		поэтому делаем цикл с lastID и добавляем в срез returnsFBO
	*/
	for hasNext {
		returns, err := ozon.NewClient(clientID, token).ReturnsList(lastID, since, to)
		if err != nil {
			return returnsWithCount, err
		}

		for _, value := range returns.Returns {
			if value.Visual.Status.SysName == "ReturnedToOzon" {
				returnsWithCount[value.Product.OfferID] += value.Product.Quantity
			}
			lastID = value.ID
		}
		// пустая страница с has_next зациклила бы запрос с тем же lastID
		hasNext = returns.HasNext && len(returns.Returns) > 0
	}

	return returnsWithCount, nil
}

// cabinetsOrders — отчёт Ozon по всем кабинетам: секции в порядке кабинетов.
type cabinetsOrders []OrdersManager

// OrdersReport — заказы и возвраты всех кабинетов за сутки day. Если не
// собрался хотя бы один кабинет, день не пишется: секции идут по порядку.
func (cc cabinetsOrders) OrdersReport(_ context.Context, day time.Time) (tradeplus.OrdersReport, error) {
	report := tradeplus.OrdersReport{Day: day}
	for _, m := range cc {
		section, err := m.section(day)
		if err != nil {
			return tradeplus.OrdersReport{}, fmt.Errorf("ozon %s: %w", m.clientID, err)
		}
		report.Sections = append(report.Sections, section)
	}
	return report, nil
}
