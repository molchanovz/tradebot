package wb_stocks_analyze

type Stock struct {
	LastChangeDate  string `json:"lastChangeDate"`
	WarehouseName   string `json:"warehouseName"`
	SupplierArticle string `json:"supplierArticle"`
	NmId            int    `json:"nmId"`
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
