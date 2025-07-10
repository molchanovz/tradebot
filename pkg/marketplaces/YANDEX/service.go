package YANDEX

import (
	"tradebot/pkg/db"
	"tradebot/pkg/marketplaces"
	"tradebot/pkg/marketplaces/OZON"
)

const (
	daysAgo = OZON.OrdersDaysAgo
)

type Service struct {
	Authorizations []marketplaces.Authorization
	SheetLink      string
}

func NewService(cabinets ...db.Cabinet) Service {
	service := Service{
		Authorizations: make([]marketplaces.Authorization, 0),
	}

	if cabinets[0].SheetLink != nil {
		service.SheetLink = *cabinets[0].SheetLink
	}

	for _, c := range cabinets {
		a := marketplaces.Authorization{
			Token: c.Key,
			Type:  c.Type,
		}
		if c.ClientID != nil {
			a.ClientId = *c.ClientID
		}

		service.Authorizations = append(service.Authorizations, a)
	}

	return service
}

func (s Service) GetOrdersAndReturnsManager() OrdersManager {
	var yandexCampaignIdFBO string
	var yandexCampaignIdFBS string
	for _, a := range s.Authorizations {
		switch a.Type {
		case "fbo":
			yandexCampaignIdFBO = a.ClientId
		case "fbs":
			yandexCampaignIdFBS = a.ClientId
		}
	}

	return NewOrdersManager(yandexCampaignIdFBO, yandexCampaignIdFBS, s.Authorizations[0].Token, s.SheetLink, daysAgo)
}

func (s Service) GetStickersFbsManager() *StickersManager {

	var yandexCampaignIdFBS string
	for _, a := range s.Authorizations {
		switch a.Type {
		case "fbs":
			yandexCampaignIdFBS = a.ClientId
		}
	}

	return NewStickersManager(yandexCampaignIdFBS, s.Authorizations[0].Token)
}
