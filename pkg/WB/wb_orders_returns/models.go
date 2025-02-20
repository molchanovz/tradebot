package wb_orders_returns

import "time"

type OrdersListALL []struct {
	Date            string  `json:"date"`
	LastChangeDate  string  `json:"lastChangeDate"`
	WarehouseName   string  `json:"warehouseName"`
	CountryName     string  `json:"countryName"`
	OblastOkrugName string  `json:"oblastOkrugName"`
	RegionName      string  `json:"regionName"`
	SupplierArticle string  `json:"supplierArticle"`
	NmId            int     `json:"nmId"`
	Barcode         string  `json:"barcode"`
	Category        string  `json:"category"`
	Subject         string  `json:"subject"`
	Brand           string  `json:"brand"`
	TechSize        string  `json:"techSize"`
	IncomeID        int     `json:"incomeID"`
	IsSupply        bool    `json:"isSupply"`
	IsRealization   bool    `json:"isRealization"`
	TotalPrice      float64 `json:"totalPrice"`
	DiscountPercent int     `json:"discountPercent"`
	Spp             int     `json:"spp"`
	FinishedPrice   float64 `json:"finishedPrice"`
	PriceWithDisc   float64 `json:"priceWithDisc"`
	IsCancel        bool    `json:"isCancel"`
	CancelDate      string  `json:"cancelDate"`
	OrderType       string  `json:"orderType"`
	Sticker         string  `json:"sticker"`
	GNumber         string  `json:"gNumber"`
	Srid            string  `json:"srid"`
}

type OrdersListFBS struct {
	Next      int `json:"next"`
	OrdersFBS []struct {
		Address struct {
			FullAddress string  `json:"fullAddress"`
			Province    string  `json:"province"`
			Area        string  `json:"area"`
			City        string  `json:"city"`
			Street      string  `json:"street"`
			Home        string  `json:"home"`
			Flat        string  `json:"flat"`
			Entrance    string  `json:"entrance"`
			Longitude   float64 `json:"longitude"`
			Latitude    float64 `json:"latitude"`
		} `json:"address"`
		ScanPrice             float64   `json:"scanPrice"`
		DeliveryType          string    `json:"deliveryType"`
		SupplyId              string    `json:"supplyId"`
		OrderUid              string    `json:"orderUid"`
		Article               string    `json:"article"`
		ColorCode             string    `json:"colorCode"`
		Rid                   string    `json:"rid"`
		CreatedAt             time.Time `json:"createdAt"`
		Offices               []string  `json:"offices"`
		Skus                  []string  `json:"skus"`
		Id                    int       `json:"id"`
		WarehouseId           int       `json:"warehouseId"`
		NmId                  int       `json:"nmId"`
		ChrtId                int       `json:"chrtId"`
		Price                 float64   `json:"price"`
		ConvertedPrice        float64   `json:"convertedPrice"`
		CurrencyCode          int       `json:"currencyCode"`
		ConvertedCurrencyCode int       `json:"convertedCurrencyCode"`
		CargoType             int       `json:"cargoType"`
		IsZeroOrder           bool      `json:"isZeroOrder"`
	} `json:"orders"`
}

type SalesAndReturns []struct {
	Date              string  `json:"date"`
	LastChangeDate    string  `json:"lastChangeDate"`
	WarehouseName     string  `json:"warehouseName"`
	CountryName       string  `json:"countryName"`
	OblastOkrugName   string  `json:"oblastOkrugName"`
	RegionName        string  `json:"regionName"`
	SupplierArticle   string  `json:"supplierArticle"`
	NmId              int     `json:"nmId"`
	Barcode           string  `json:"barcode"`
	Category          string  `json:"category"`
	Subject           string  `json:"subject"`
	Brand             string  `json:"brand"`
	TechSize          string  `json:"techSize"`
	IncomeID          int     `json:"incomeID"`
	IsSupply          bool    `json:"isSupply"`
	IsRealization     bool    `json:"isRealization"`
	TotalPrice        float64 `json:"totalPrice"`
	DiscountPercent   int     `json:"discountPercent"`
	Spp               int     `json:"spp"`
	PaymentSaleAmount int     `json:"paymentSaleAmount"`
	ForPay            float64 `json:"forPay"`
	FinishedPrice     float64 `json:"finishedPrice"`
	PriceWithDisc     float64 `json:"priceWithDisc"`
	SaleID            string  `json:"saleID"`
	OrderType         string  `json:"orderType"`
	Sticker           string  `json:"sticker"`
	GNumber           string  `json:"gNumber"`
	Srid              string  `json:"srid"`
}

type OrdersWithStatuses struct {
	Orders []struct {
		Id             int    `json:"id"`
		SupplierStatus string `json:"supplierStatus"`
		WbStatus       string `json:"wbStatus"`
	} `json:"orders"`
}
