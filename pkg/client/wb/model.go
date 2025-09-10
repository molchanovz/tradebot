package wb

import "time"

type OrderWB struct {
	OrderUID              string   `json:"orderUid"`
	Article               string   `json:"article"`
	ColorCode             string   `json:"colorCode"`
	RID                   string   `json:"rid"`
	CreatedAt             string   `json:"createdAt"`
	Offices               []string `json:"offices"`
	SKUs                  []string `json:"skus"`
	ID                    int      `json:"id"`
	WarehouseID           int      `json:"warehouseId"`
	NmID                  int      `json:"nmId"`
	ChrtID                int      `json:"chrtId"`
	Price                 int      `json:"price"`
	ConvertedPrice        int      `json:"convertedPrice"`
	CurrencyCode          int      `json:"currencyCode"`
	ConvertedCurrencyCode int      `json:"convertedCurrencyCode"`
	CargoType             int      `json:"cargoType"`
	IsZeroOrder           bool     `json:"isZeroOrder"`
}
type Orders struct {
	Orders []OrderWB `json:"orders"`
}

type StickerWB struct {
	Stickers []struct {
		OrderID int    `json:"orderId"`
		PartA   string `json:"partA"`
		PartB   string `json:"partB"`
		Barcode string `json:"barcode"`
		File    string `json:"file"`
	} `json:"stickers"`
}

type OrdersListALL []struct {
	Date            string  `json:"date"`
	LastChangeDate  string  `json:"lastChangeDate"`
	WarehouseName   string  `json:"warehouseName"`
	WarehouseType   string  `json:"warehouseType"`
	CountryName     string  `json:"countryName"`
	OblastOkrugName string  `json:"oblastOkrugName"`
	RegionName      string  `json:"regionName"`
	SupplierArticle string  `json:"supplierArticle"`
	NmID            int     `json:"nmId"`
	Barcode         string  `json:"barcode"`
	Category        string  `json:"category"`
	Subject         string  `json:"subject"`
	Brand           string  `json:"brand"`
	TechSize        string  `json:"techSize"`
	IncomeID        int     `json:"incomeID"`
	IsSupply        bool    `json:"isSupply"`
	IsRealization   bool    `json:"isRealization"`
	TotalPrice      int     `json:"totalPrice"`
	DiscountPercent int     `json:"discountPercent"`
	Spp             int     `json:"spp"`
	FinishedPrice   float64 `json:"finishedPrice"`
	PriceWithDisc   float64 `json:"priceWithDisc"`
	IsCancel        bool    `json:"isCancel"`
	CancelDate      string  `json:"cancelDate"`
	Sticker         string  `json:"sticker"`
	GNumber         string  `json:"gNumber"`
	SrID            string  `json:"srid"`
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
		SupplyID              string    `json:"supplyId"`
		OrderUID              string    `json:"orderUid"`
		Article               string    `json:"article"`
		ColorCode             string    `json:"colorCode"`
		RID                   string    `json:"rid"`
		CreatedAt             time.Time `json:"createdAt"`
		Offices               []string  `json:"offices"`
		Skus                  []string  `json:"skus"`
		ID                    int       `json:"id"`
		WarehouseID           int       `json:"warehouseId"`
		NmID                  int       `json:"nmId"`
		ChrtID                int       `json:"chrtId"`
		Price                 float64   `json:"price"`
		ConvertedPrice        float64   `json:"convertedPrice"`
		CurrencyCode          int       `json:"currencyCode"`
		ConvertedCurrencyCode int       `json:"convertedCurrencyCode"`
		CargoType             int       `json:"cargoType"`
		IsZeroOrder           bool      `json:"isZeroOrder"`
	} `json:"orders"`
}

type SalesReturns []struct {
	Date              string  `json:"date"`
	LastChangeDate    string  `json:"lastChangeDate"`
	WarehouseName     string  `json:"warehouseName"`
	CountryName       string  `json:"countryName"`
	OblastOkrugName   string  `json:"oblastOkrugName"`
	RegionName        string  `json:"regionName"`
	SupplierArticle   string  `json:"supplierArticle"`
	NmID              int     `json:"nmId"`
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
	SrID              string  `json:"srid"`
}

type OrdersWithStatuses struct {
	Orders []struct {
		ID             int    `json:"id"`
		SupplierStatus string `json:"supplierStatus"`
		WbStatus       string `json:"wbStatus"`
	} `json:"orders"`
}

type Stock struct {
	LastChangeDate  string `json:"lastChangeDate"`
	WarehouseName   string `json:"warehouseName"`
	SupplierArticle string `json:"supplierArticle"`
	NmID            int    `json:"nmId"`
	Barcode         string `json:"barcode"`
	Quantity        int    `json:"quantity"`
	InWayToClient   int    `json:"inWayToClient"`
	InWayFromClient int    `json:"inWayFromClient"`
	QuantityFull    int    `json:"quantityFull"`
	Category        string `json:"category"`
	Subject         string `json:"subject"`
	Brand           string `json:"brand"`
	TechSize        string `json:"techSize"`
	Price           int    `json:"Price"`
	Discount        int    `json:"Discount"`
	IsSupply        bool   `json:"isSupply"`
	IsRealization   bool   `json:"isRealization"`
	SCCode          string `json:"SCCode"`
}

type ReturnList struct {
	Report []struct {
		Barcode          string `json:"barcode"`
		Brand            string `json:"brand"`
		CompletedDt      string `json:"completedDt"`
		DstOfficeAddress string `json:"dstOfficeAddress"`
		DstOfficeId      int    `json:"dstOfficeId"`
		ExpiredDt        string `json:"expiredDt"`
		IsStatusActive   int    `json:"isStatusActive"`
		NmId             int    `json:"nmId"`
		OrderDt          string `json:"orderDt"`
		OrderId          int    `json:"orderId"`
		ReadyToReturnDt  string `json:"readyToReturnDt"`
		Reason           string `json:"reason"`
		ReturnType       string `json:"returnType"`
		ShkId            int64  `json:"shkId"`
		Srid             string `json:"srid"`
		Status           string `json:"status"`
		StickerId        string `json:"stickerId"`
		SubjectName      string `json:"subjectName"`
		TechSize         string `json:"techSize"`
	} `json:"report"`
}

type CardList struct {
	Cards []struct {
		NmID        int    `json:"nmID"`
		ImtID       int    `json:"imtID"`
		NmUUID      string `json:"nmUUID"`
		SubjectID   int    `json:"subjectID"`
		SubjectName string `json:"subjectName"`
		VendorCode  string `json:"vendorCode"`
		Brand       string `json:"brand"`
		Title       string `json:"title"`
		Description string `json:"description"`
		NeedKiz     bool   `json:"needKiz"`
		Dimensions  struct {
			Width        int     `json:"width"`
			Height       int     `json:"height"`
			Length       int     `json:"length"`
			WeightBrutto float64 `json:"weightBrutto"`
			IsValid      bool    `json:"isValid"`
		} `json:"dimensions"`
		Characteristics []struct {
			Id    int         `json:"id"`
			Name  string      `json:"name"`
			Value interface{} `json:"value"`
		} `json:"characteristics"`
		Sizes []struct {
			ChrtID   int      `json:"chrtID"`
			TechSize string   `json:"techSize"`
			WbSize   string   `json:"wbSize"`
			Skus     []string `json:"skus"`
		} `json:"sizes"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	} `json:"cards"`
	Cursor struct {
		UpdatedAt time.Time `json:"updatedAt"`
		NmID      int       `json:"nmID"`
		Total     int       `json:"total"`
	} `json:"cursor"`
}
