package model

import "time"

type OrdersFbo struct {
	Status string `json:"status"`
	Result struct {
		Orders []struct {
			Id               int64     `json:"id"`
			CreationDate     string    `json:"creationDate"`
			StatusUpdateDate time.Time `json:"statusUpdateDate"`
			Status           string    `json:"status"`
			PartnerOrderId   string    `json:"partnerOrderId"`
			PaymentType      string    `json:"paymentType"`
			Fake             bool      `json:"fake"`
			DeliveryRegion   struct {
				Id   int    `json:"id"`
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
					Id   int    `json:"id"`
					Name string `json:"name"`
				} `json:"warehouse"`
				Details []interface{} `json:"details"`
				CisList []interface{} `json:"cisList"`
				BidFee  int           `json:"bidFee,omitempty"`
			} `json:"items"`
			Payments []struct {
				Id     string  `json:"id"`
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
