package yandex

import "time"

type OrdersFbo struct {
	Status string `json:"status"`
	Result struct {
		Orders []struct {
			ID               int64     `json:"id"`
			CreationDate     string    `json:"creationDate"`
			StatusUpdateDate time.Time `json:"statusUpdateDate"`
			Status           string    `json:"status"`
			PartnerOrderID   string    `json:"partnerOrderId"`
			PaymentType      string    `json:"paymentType"`
			Fake             bool      `json:"fake"`
			DeliveryRegion   struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"deliveryRegion"`
			Items []struct {
				OfferName string `json:"offerName"`
				MarketSku int64  `json:"marketSku"`
				ShopSku   string `json:"shopSku"`
				Count     int    `json:"count"`
				Prices    []struct {
					Type        string  `json:"type"`
					CostPerItem float64 `json:"costPerItem"`
					Total       float64 `json:"total"`
				} `json:"prices"`
				Warehouse struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"warehouse"`
				Details []interface{} `json:"details"`
				CisList []interface{} `json:"cisList"`
				BidFee  int           `json:"bidFee,omitempty"`
			} `json:"items"`
			Payments []struct {
				ID     string  `json:"id"`
				Date   string  `json:"date"`
				Type   string  `json:"type"`
				Source string  `json:"source"`
				Total  float64 `json:"total"`
			} `json:"payments"`
			Commissions []struct {
				Type   string  `json:"type"`
				Actual float64 `json:"actual"`
			} `json:"commissions"`
			Subsidies []struct {
				OperationType string  `json:"operationType"`
				Type          string  `json:"type"`
				Amount        float64 `json:"amount"`
			} `json:"subsidies,omitempty"`
			BuyerType string `json:"buyerType"`
			Currency  string `json:"currency"`
		} `json:"orders"`
		Paging struct {
			NextPageToken string `json:"nextPageToken"`
		} `json:"paging"`
	} `json:"result"`
}

type Shipment struct {
	Status string `json:"status"`
	Result struct {
		ID               int       `json:"id"`
		PlanIntervalFrom time.Time `json:"planIntervalFrom"`
		PlanIntervalTo   time.Time `json:"planIntervalTo"`
		ShipmentType     string    `json:"shipmentType"`
		Warehouse        struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"warehouse"`
		WarehouseTo struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"warehouseTo"`
		ExternalID      string `json:"externalId"`
		DeliveryService struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"deliveryService"`
		PalletsCount struct {
			Planned int `json:"planned"`
		} `json:"palletsCount"`
		OrderIds      []int64 `json:"orderIds"`
		DraftCount    int     `json:"draftCount"`
		PlannedCount  int     `json:"plannedCount"`
		FactCount     int     `json:"factCount"`
		CurrentStatus struct {
			Status      string    `json:"status"`
			Description string    `json:"description"`
			UpdateTime  time.Time `json:"updateTime"`
		} `json:"currentStatus"`
		AvailableActions []string `json:"availableActions"`
	} `json:"result"`
}

type Order struct {
	Order struct {
		ID                            int64   `json:"id"`
		Status                        string  `json:"status"`
		SubStatus                     string  `json:"substatus"`
		CreationDate                  string  `json:"creationDate"`
		UpdatedAt                     string  `json:"updatedAt"`
		Currency                      string  `json:"currency"`
		ItemsTotal                    float64 `json:"itemsTotal"`
		DeliveryTotal                 float64 `json:"deliveryTotal"`
		BuyerItemsTotal               float64 `json:"buyerItemsTotal"`
		BuyerTotal                    float64 `json:"buyerTotal"`
		BuyerItemsTotalBeforeDiscount float64 `json:"buyerItemsTotalBeforeDiscount"`
		BuyerTotalBeforeDiscount      float64 `json:"buyerTotalBeforeDiscount"`
		PaymentType                   string  `json:"paymentType"`
		PaymentMethod                 string  `json:"paymentMethod"`
		Fake                          bool    `json:"fake"`
		Items                         []struct {
			ID                       int     `json:"id"`
			OfferID                  string  `json:"offerId"`
			OfferName                string  `json:"offerName"`
			Price                    float64 `json:"price"`
			BuyerPrice               float64 `json:"buyerPrice"`
			BuyerPriceBeforeDiscount float64 `json:"buyerPriceBeforeDiscount"`
			PriceBeforeDiscount      float64 `json:"priceBeforeDiscount"`
			Count                    int     `json:"count"`
			Vat                      string  `json:"vat"`
			ShopSku                  string  `json:"shopSku"`
			Subsidy                  float64 `json:"subsidy"`
			PartnerWarehouseID       string  `json:"partnerWarehouseId"`
			Promos                   []struct {
				Type    string  `json:"type"`
				Subsidy float64 `json:"subsidy"`
			} `json:"promos"`
			Subsidies []struct {
				Type   string  `json:"type"`
				Amount float64 `json:"amount"`
			} `json:"subsidies"`
		} `json:"items"`
		Subsidies []struct {
			Type   string  `json:"type"`
			Amount float64 `json:"amount"`
		} `json:"subsidies"`
		Delivery struct {
			Type                string  `json:"type"`
			ServiceName         string  `json:"serviceName"`
			Price               float64 `json:"price"`
			DeliveryPartnerType string  `json:"deliveryPartnerType"`
			Dates               struct {
				FromDate string `json:"fromDate"`
				ToDate   string `json:"toDate"`
				FromTime string `json:"fromTime"`
				ToTime   string `json:"toTime"`
			} `json:"dates"`
			Region struct {
				ID     int    `json:"id"`
				Name   string `json:"name"`
				Type   string `json:"type"`
				Parent struct {
					ID     int    `json:"id"`
					Name   string `json:"name"`
					Type   string `json:"type"`
					Parent struct {
						ID     int    `json:"id"`
						Name   string `json:"name"`
						Type   string `json:"type"`
						Parent struct {
							ID   int    `json:"id"`
							Name string `json:"name"`
							Type string `json:"type"`
						} `json:"parent"`
					} `json:"parent"`
				} `json:"parent"`
			} `json:"region"`
			Address struct {
				Country  string `json:"country"`
				Postcode string `json:"postcode"`
				City     string `json:"city"`
				Street   string `json:"street"`
				House    string `json:"house"`
				Block    string `json:"block"`
				Gps      struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"gps"`
			} `json:"address"`
			DeliveryServiceID int     `json:"deliveryServiceId"`
			LiftPrice         float64 `json:"liftPrice"`
			OutletCode        string  `json:"outletCode"`
			Shipments         []struct {
				ID           int    `json:"id"`
				ShipmentDate string `json:"shipmentDate"`
				Boxes        []struct {
					ID           int    `json:"id"`
					FulfilmentID string `json:"fulfilmentId"`
				} `json:"boxes"`
			} `json:"shipments"`
		} `json:"delivery"`
		Buyer struct {
			Type string `json:"type"`
		} `json:"buyer"`
		TaxSystem       string `json:"taxSystem"`
		CancelRequested bool   `json:"cancelRequested"`
	} `json:"order"`
}

type Items []struct {
	ID                       int     `json:"id"`
	OfferID                  string  `json:"offerId"`
	OfferName                string  `json:"offerName"`
	Price                    float64 `json:"price"`
	BuyerPrice               float64 `json:"buyerPrice"`
	BuyerPriceBeforeDiscount float64 `json:"buyerPriceBeforeDiscount"`
	PriceBeforeDiscount      float64 `json:"priceBeforeDiscount"`
	Count                    int     `json:"count"`
	Vat                      string  `json:"vat"`
	ShopSku                  string  `json:"shopSku"`
	Subsidy                  float64 `json:"subsidy"`
	PartnerWarehouseID       string  `json:"partnerWarehouseId"`
	Promos                   []struct {
		Type    string  `json:"type"`
		Subsidy float64 `json:"subsidy"`
	} `json:"promos"`
	Subsidies []struct {
		Type   string  `json:"type"`
		Amount float64 `json:"amount"`
	} `json:"subsidies"`
}
