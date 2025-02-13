package API

import "time"

type ReturnFBO struct {
	ReturnID                   int    `json:"return_id"`
	ID                         int    `json:"id"`
	SKU                        int    `json:"sku"`
	CompanyID                  int    `json:"company_id"`
	PostingNumber              string `json:"posting_number"`
	AcceptedFromCustomerMoment string `json:"accepted_from_customer_moment"`
	ReturnReasonName           string `json:"return_reason_name"`
	IsOpened                   bool   `json:"is_opened"`
	StatusName                 string `json:"status_name"`
	ReturnedToOzonMoment       string `json:"returned_to_ozon_moment"`
	CurrentPlaceName           string `json:"current_place_name"`
	DstPlaceName               string `json:"dst_place_name"`
}

type ReturnsFBO struct {
	Returns []ReturnFBO `json:"returns"`
	LastID  int         `json:"last_id"`
}

type PostingFBO struct {
	Result struct {
		OrderId        int64     `json:"order_id"`
		OrderNumber    string    `json:"order_number"`
		PostingNumber  string    `json:"posting_number"`
		Status         string    `json:"status"`
		CancelReasonId int       `json:"cancel_reason_id"`
		CreatedAt      time.Time `json:"created_at"`
		InProcessAt    time.Time `json:"in_process_at"`
		Products       []struct {
			Sku          int           `json:"sku"`
			Name         string        `json:"name"`
			Quantity     int           `json:"quantity"`
			OfferId      string        `json:"offer_id"`
			Price        string        `json:"price"`
			DigitalCodes []interface{} `json:"digital_codes"`
			CurrencyCode string        `json:"currency_code"`
		} `json:"products"`
		AnalyticsData  interface{}   `json:"analytics_data"`
		FinancialData  interface{}   `json:"financial_data"`
		AdditionalData []interface{} `json:"additional_data"`
	} `json:"result"`
}

type ReturnFBS struct {
	ClearingId               int     `json:"clearing_id"`
	Commission               float64 `json:"commission"`
	CommissionPercent        float64 `json:"commission_percent"`
	ExemplarId               int     `json:"exemplar_id"`
	Id                       int     `json:"id"`
	IsMoving                 bool    `json:"is_moving"`
	IsOpened                 bool    `json:"is_opened"`
	LastFreeWaitingDay       string  `json:"last_free_waiting_day"`
	PlaceId                  int     `json:"place_id"`
	MovingToPlaceName        string  `json:"moving_to_place_name"`
	PickingAmount            int     `json:"picking_amount"`
	PostingNumber            string  `json:"posting_number"`
	PickingTag               string  `json:"picking_tag"`
	Price                    float64 `json:"price"`
	PriceWithoutCommission   float64 `json:"price_without_commission"`
	ProductId                int     `json:"product_id"`
	ProductName              string  `json:"product_name"`
	Quantity                 int     `json:"quantity"`
	ReturnBarcode            string  `json:"return_barcode"`
	ReturnClearingId         int     `json:"return_clearing_id"`
	ReturnDate               string  `json:"return_date"`
	ReturnReasonName         string  `json:"return_reason_name"`
	WaitingForSellerDateTime string  `json:"waiting_for_seller_date_time"`
	ReturnedToSellerDateTime string  `json:"returned_to_seller_date_time"`
	WaitingForSellerDays     int     `json:"waiting_for_seller_days"`
	ReturnsKeepingCost       int     `json:"returns_keeping_cost"`
	Sku                      int     `json:"sku"`
	Status                   string  `json:"status"`
}

type ReturnsFBS struct {
	Returns []ReturnFBS `json:"returns"`
	LastID  int         `json:"last_id"`
}

type PostingFBS struct {
	Result struct {
		PostingNumber  string `json:"posting_number"`
		OrderId        int    `json:"order_id"`
		OrderNumber    string `json:"order_number"`
		Status         string `json:"status"`
		Substatus      string `json:"substatus"`
		DeliveryMethod struct {
			Id            int64  `json:"id"`
			Name          string `json:"name"`
			WarehouseId   int64  `json:"warehouse_id"`
			Warehouse     string `json:"warehouse"`
			TplProviderId int    `json:"tpl_provider_id"`
			TplProvider   string `json:"tpl_provider"`
		} `json:"delivery_method"`
		TrackingNumber     string      `json:"tracking_number"`
		TplIntegrationType string      `json:"tpl_integration_type"`
		InProcessAt        time.Time   `json:"in_process_at"`
		ShipmentDate       time.Time   `json:"shipment_date"`
		DeliveringDate     interface{} `json:"delivering_date"`
		ProviderStatus     string      `json:"provider_status"`
		DeliveryPrice      string      `json:"delivery_price"`
		Cancellation       struct {
			CancelReasonId           int    `json:"cancel_reason_id"`
			CancelReason             string `json:"cancel_reason"`
			CancellationType         string `json:"cancellation_type"`
			CancelledAfterShip       bool   `json:"cancelled_after_ship"`
			AffectCancellationRating bool   `json:"affect_cancellation_rating"`
			CancellationInitiator    string `json:"cancellation_initiator"`
		} `json:"cancellation"`
		Customer  interface{} `json:"customer"`
		Addressee interface{} `json:"addressee"`
		Products  []struct {
			CurrencyCode  string        `json:"currency_code"`
			Price         string        `json:"price"`
			OfferId       string        `json:"offer_id"`
			Name          string        `json:"name"`
			Sku           int           `json:"sku"`
			Quantity      int           `json:"quantity"`
			JwUin         []interface{} `json:"jw_uin"`
			MandatoryMark []interface{} `json:"mandatory_mark"`
			Dimensions    struct {
				Height string `json:"height"`
				Length string `json:"length"`
				Weight string `json:"weight"`
				Width  string `json:"width"`
			} `json:"dimensions"`
		} `json:"products"`
		Barcodes       interface{}   `json:"barcodes"`
		AnalyticsData  interface{}   `json:"analytics_data"`
		FinancialData  interface{}   `json:"financial_data"`
		AdditionalData []interface{} `json:"additional_data"`
		IsExpress      bool          `json:"is_express"`
		Requirements   struct {
			ProductsRequiringGtd     []interface{} `json:"products_requiring_gtd"`
			ProductsRequiringCountry []interface{} `json:"products_requiring_country"`
			ProductsRequiringJwn     []interface{} `json:"products_requiring_jwn"`
		} `json:"requirements"`
		ProductExemplars interface{} `json:"product_exemplars"`
	} `json:"result"`
}

type ReportResponse struct {
	Result struct {
		Code string `json:"code"`
	} `json:"result"`
}

type PostingsList_FBS struct {
	Result struct {
		PostingsFBS []struct {
			PostingNumber  string `json:"posting_number"`
			OrderId        int    `json:"order_id"`
			OrderNumber    string `json:"order_number"`
			Status         string `json:"status"`
			Substatus      string `json:"substatus"`
			DeliveryMethod struct {
				Id            int64  `json:"id"`
				Name          string `json:"name"`
				WarehouseId   int64  `json:"warehouse_id"`
				Warehouse     string `json:"warehouse"`
				TplProviderId int    `json:"tpl_provider_id"`
				TplProvider   string `json:"tpl_provider"`
			} `json:"delivery_method"`
			TrackingNumber     string      `json:"tracking_number"`
			TplIntegrationType string      `json:"tpl_integration_type"`
			InProcessAt        time.Time   `json:"in_process_at"`
			ShipmentDate       time.Time   `json:"shipment_date"`
			DeliveringDate     interface{} `json:"delivering_date"`
			Cancellation       struct {
				CancelReasonId           int    `json:"cancel_reason_id"`
				CancelReason             string `json:"cancel_reason"`
				CancellationType         string `json:"cancellation_type"`
				CancelledAfterShip       bool   `json:"cancelled_after_ship"`
				AffectCancellationRating bool   `json:"affect_cancellation_rating"`
				CancellationInitiator    string `json:"cancellation_initiator"`
			} `json:"cancellation"`
			Customer interface{} `json:"customer"`
			Products []struct {
				Price         string        `json:"price"`
				CurrencyCode  string        `json:"currency_code"`
				OfferId       string        `json:"offer_id"`
				Name          string        `json:"name"`
				Sku           int           `json:"sku"`
				Quantity      int           `json:"quantity"`
				MandatoryMark []interface{} `json:"mandatory_mark"`
			} `json:"products"`
			Addressee     interface{} `json:"addressee"`
			Barcodes      interface{} `json:"barcodes"`
			AnalyticsData interface{} `json:"analytics_data"`

			FinancialData struct {
				Products []struct {
					CommissionAmount     float64     `json:"commission_amount"`
					CommissionPercent    int         `json:"commission_percent"`
					Payout               float64     `json:"payout"`
					ProductId            int         `json:"product_id"`
					OldPrice             float64     `json:"old_price"`
					Price                float64     `json:"price"`
					TotalDiscountValue   float64     `json:"total_discount_value"`
					TotalDiscountPercent float64     `json:"total_discount_percent"`
					Actions              []string    `json:"actions"`
					Picking              interface{} `json:"picking"`
					Quantity             int         `json:"quantity"`
					ClientPrice          string      `json:"client_price"`
					ItemServices         struct {
						MarketplaceServiceItemFulfillment                int     `json:"marketplace_service_item_fulfillment"`
						MarketplaceServiceItemPickup                     int     `json:"marketplace_service_item_pickup"`
						MarketplaceServiceItemDropoffPvz                 int     `json:"marketplace_service_item_dropoff_pvz"`
						MarketplaceServiceItemDropoffSc                  int     `json:"marketplace_service_item_dropoff_sc"`
						MarketplaceServiceItemDropoffFf                  int     `json:"marketplace_service_item_dropoff_ff"`
						MarketplaceServiceItemDirectFlowTrans            int     `json:"marketplace_service_item_direct_flow_trans"`
						MarketplaceServiceItemReturnFlowTrans            int     `json:"marketplace_service_item_return_flow_trans"`
						MarketplaceServiceItemDelivToCustomer            float64 `json:"marketplace_service_item_deliv_to_customer"`
						MarketplaceServiceItemReturnNotDelivToCustomer   int     `json:"marketplace_service_item_return_not_deliv_to_customer"`
						MarketplaceServiceItemReturnPartGoodsCustomer    int     `json:"marketplace_service_item_return_part_goods_customer"`
						MarketplaceServiceItemReturnAfterDelivToCustomer int     `json:"marketplace_service_item_return_after_deliv_to_customer"`
					} `json:"item_services"`
					CurrencyCode string `json:"currency_code"`
				} `json:"products"`
				PostingServices struct {
					MarketplaceServiceItemFulfillment                int `json:"marketplace_service_item_fulfillment"`
					MarketplaceServiceItemPickup                     int `json:"marketplace_service_item_pickup"`
					MarketplaceServiceItemDropoffPvz                 int `json:"marketplace_service_item_dropoff_pvz"`
					MarketplaceServiceItemDropoffSc                  int `json:"marketplace_service_item_dropoff_sc"`
					MarketplaceServiceItemDropoffFf                  int `json:"marketplace_service_item_dropoff_ff"`
					MarketplaceServiceItemDirectFlowTrans            int `json:"marketplace_service_item_direct_flow_trans"`
					MarketplaceServiceItemReturnFlowTrans            int `json:"marketplace_service_item_return_flow_trans"`
					MarketplaceServiceItemDelivToCustomer            int `json:"marketplace_service_item_deliv_to_customer"`
					MarketplaceServiceItemReturnNotDelivToCustomer   int `json:"marketplace_service_item_return_not_deliv_to_customer"`
					MarketplaceServiceItemReturnPartGoodsCustomer    int `json:"marketplace_service_item_return_part_goods_customer"`
					MarketplaceServiceItemReturnAfterDelivToCustomer int `json:"marketplace_service_item_return_after_deliv_to_customer"`
				} `json:"posting_services"`
				ClusterFrom string `json:"cluster_from"`
				ClusterTo   string `json:"cluster_to"`
			} `json:"financial_data"`
			IsExpress    bool `json:"is_express"`
			Requirements struct {
				ProductsRequiringGtd           []interface{} `json:"products_requiring_gtd"`
				ProductsRequiringCountry       []interface{} `json:"products_requiring_country"`
				ProductsRequiringMandatoryMark []interface{} `json:"products_requiring_mandatory_mark"`
				ProductsRequiringJwn           []interface{} `json:"products_requiring_jwn"`
			} `json:"requirements"`
		} `json:"postings"`
		HasNext bool `json:"has_next"`
	} `json:"result"`
}

type PostingsList_FBO struct {
	Result []struct {
		OrderId        int       `json:"order_id"`
		OrderNumber    string    `json:"order_number"`
		PostingNumber  string    `json:"posting_number"`
		Status         string    `json:"status"`
		CancelReasonId int       `json:"cancel_reason_id"`
		CreatedAt      time.Time `json:"created_at"`
		InProcessAt    time.Time `json:"in_process_at"`
		Products       []struct {
			Sku          int           `json:"sku"`
			Name         string        `json:"name"`
			Quantity     int           `json:"quantity"`
			OfferId      string        `json:"offer_id"`
			Price        string        `json:"price"`
			DigitalCodes []interface{} `json:"digital_codes"`
			CurrencyCode string        `json:"currency_code"`
		} `json:"products"`
		AnalyticsData struct {
			Region               string `json:"region"`
			City                 string `json:"city"`
			DeliveryType         string `json:"delivery_type"`
			IsPremium            bool   `json:"is_premium"`
			PaymentTypeGroupName string `json:"payment_type_group_name"`
			WarehouseId          int64  `json:"warehouse_id"`
			WarehouseName        string `json:"warehouse_name"`
			IsLegal              bool   `json:"is_legal"`
		} `json:"analytics_data"`

		FinancialData struct {
			Products []struct {
				CommissionAmount     float64     `json:"commission_amount"`
				CommissionPercent    int         `json:"commission_percent"`
				Payout               float64     `json:"payout"`
				ProductId            int         `json:"product_id"`
				OldPrice             float64     `json:"old_price"`
				Price                float64     `json:"price"`
				TotalDiscountValue   float64     `json:"total_discount_value"`
				TotalDiscountPercent float64     `json:"total_discount_percent"`
				Actions              []string    `json:"actions"`
				Picking              interface{} `json:"picking"`
				ClientPrice          string      `json:"client_price"`
				ItemServices         struct {
					MarketplaceServiceItemFulfillment                int     `json:"marketplace_service_item_fulfillment"`
					MarketplaceServiceItemPickup                     int     `json:"marketplace_service_item_pickup"`
					MarketplaceServiceItemDropoffPvz                 int     `json:"marketplace_service_item_dropoff_pvz"`
					MarketplaceServiceItemDropoffSc                  int     `json:"marketplace_service_item_dropoff_sc"`
					MarketplaceServiceItemDropoffFf                  int     `json:"marketplace_service_item_dropoff_ff"`
					MarketplaceServiceItemDirectFlowTrans            int     `json:"marketplace_service_item_direct_flow_trans"`
					MarketplaceServiceItemReturnFlowTrans            int     `json:"marketplace_service_item_return_flow_trans"`
					MarketplaceServiceItemDelivToCustomer            float64 `json:"marketplace_service_item_deliv_to_customer"`
					MarketplaceServiceItemReturnNotDelivToCustomer   int     `json:"marketplace_service_item_return_not_deliv_to_customer"`
					MarketplaceServiceItemReturnPartGoodsCustomer    int     `json:"marketplace_service_item_return_part_goods_customer"`
					MarketplaceServiceItemReturnAfterDelivToCustomer int     `json:"marketplace_service_item_return_after_deliv_to_customer"`
				} `json:"item_services"`
				CurrencyCode string `json:"currency_code"`
			} `json:"products"`
			PostingServices struct {
				MarketplaceServiceItemFulfillment                int `json:"marketplace_service_item_fulfillment"`
				MarketplaceServiceItemPickup                     int `json:"marketplace_service_item_pickup"`
				MarketplaceServiceItemDropoffPvz                 int `json:"marketplace_service_item_dropoff_pvz"`
				MarketplaceServiceItemDropoffSc                  int `json:"marketplace_service_item_dropoff_sc"`
				MarketplaceServiceItemDropoffFf                  int `json:"marketplace_service_item_dropoff_ff"`
				MarketplaceServiceItemDirectFlowTrans            int `json:"marketplace_service_item_direct_flow_trans"`
				MarketplaceServiceItemReturnFlowTrans            int `json:"marketplace_service_item_return_flow_trans"`
				MarketplaceServiceItemDelivToCustomer            int `json:"marketplace_service_item_deliv_to_customer"`
				MarketplaceServiceItemReturnNotDelivToCustomer   int `json:"marketplace_service_item_return_not_deliv_to_customer"`
				MarketplaceServiceItemReturnPartGoodsCustomer    int `json:"marketplace_service_item_return_part_goods_customer"`
				MarketplaceServiceItemReturnAfterDelivToCustomer int `json:"marketplace_service_item_return_after_deliv_to_customer"`
			} `json:"posting_services"`
			ClusterFrom string `json:"cluster_from"`
			ClusterTo   string `json:"cluster_to"`
		} `json:"financial_data"`
		AdditionalData []interface{} `json:"additional_data"`
	} `json:"result"`
}

type StocksList struct {
	Result struct {
		Rows []struct {
			Sku              int      `json:"sku"`
			WarehouseName    string   `json:"warehouse_name"`
			ItemCode         string   `json:"item_code"`
			ItemName         string   `json:"item_name"`
			PromisedAmount   int      `json:"promised_amount"`
			FreeToSellAmount int      `json:"free_to_sell_amount"`
			ReservedAmount   int      `json:"reserved_amount"`
			Idc              *float64 `json:"idc"`
		} `json:"rows"`
	} `json:"result"`
}

type ClustersList struct {
	Clusters []struct {
		LogisticClusters []struct {
			Warehouses []struct {
				WarehouseId int64  `json:"warehouse_id"`
				Type        string `json:"type"`
				Name        string `json:"name"`
			} `json:"warehouses"`
		} `json:"logistic_clusters"`
		Id   int    `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"clusters"`
}
