package yandex

import (
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/ozon"
)

const (
	daysAgo = ozon.OrdersDaysAgo
)

type Service struct {
	Authorizations []tradeplus.Authorization
	SheetLink      string
}

func NewService(cabinets ...tradeplus.Cabinet) Service {
	service := Service{
		Authorizations: make([]tradeplus.Authorization, 0),
	}

	if cabinets[0].SheetLink != nil {
		service.SheetLink = *cabinets[0].SheetLink
	}

	for _, c := range cabinets {
		a := tradeplus.Authorization{
			Token: c.Key,
			Type:  c.Type,
		}
		if c.ClientID != nil {
			a.ClientID = *c.ClientID
		}

		service.Authorizations = append(service.Authorizations, a)
	}

	return service
}

func (s Service) GetOrdersAndReturnsManager() OrdersManager {
	var yandexCampaignIDFBO string
	var yandexCampaignIDFBS string
	for _, a := range s.Authorizations {
		switch a.Type {
		case "fbo":
			yandexCampaignIDFBO = a.ClientID
		case "fbs":
			yandexCampaignIDFBS = a.ClientID
		}
	}

	return NewOrdersManager(yandexCampaignIDFBO, yandexCampaignIDFBS, s.Authorizations[0].Token, s.SheetLink, daysAgo)
}

func (s Service) GetStickersFbsManager() *StickersManager {
	var yandexCampaignIDFBS string
	for _, a := range s.Authorizations {
		if a.Type == "fbs" {
			yandexCampaignIDFBS = a.ClientID
		}
	}

	return NewStickersManager(yandexCampaignIDFBS, s.Authorizations[0].Token)
}
