package wb

import (
	"time"
	"tradebot/pkg/client/wb"
)

//go:generate colgen

type Card struct {
	NmID        int
	ImtID       int
	NmUUID      string
	SubjectID   int
	SubjectName string
	VendorCode  string
	Brand       string
	Title       string
	Description string
	NeedKiz     bool
	Dimensions  struct {
		Width        int
		Height       int
		Length       int
		WeightBrutto float64
		IsValid      bool
	}
	Characteristics []struct {
		Id    int
		Name  string
		Value interface{}
	}
	Sizes []struct {
		ChrtID   int
		TechSize string
		WbSize   string
		Skus     []string
	}
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCardList(in *wb.CardList) Cards {
	if in == nil {
		return nil
	}

	cards := make([]Card, 0, len(in.Cards))
	for i := range in.Cards {
		c := Card{
			NmID:            in.Cards[i].NmID,
			ImtID:           in.Cards[i].ImtID,
			NmUUID:          in.Cards[i].NmUUID,
			SubjectID:       in.Cards[i].SubjectID,
			SubjectName:     in.Cards[i].SubjectName,
			VendorCode:      in.Cards[i].VendorCode,
			Brand:           in.Cards[i].Brand,
			Title:           in.Cards[i].Title,
			Description:     in.Cards[i].Description,
			NeedKiz:         in.Cards[i].NeedKiz,
			Characteristics: nil,
			Sizes:           nil,
			CreatedAt:       time.Time{},
			UpdatedAt:       time.Time{},
		}
		c.Dimensions.Height = in.Cards[i].Dimensions.Height
		c.Dimensions.Length = in.Cards[i].Dimensions.Length
		c.Dimensions.Width = in.Cards[i].Dimensions.Width
		c.Dimensions.IsValid = in.Cards[i].Dimensions.IsValid
		c.Dimensions.WeightBrutto = in.Cards[i].Dimensions.WeightBrutto

		cards = append(cards, c)
	}

	return cards
}

type Return struct {
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
}

func NewReturns(in *wb.ReturnList) []Return {
	if in == nil {
		return nil
	}

	returns := make([]Return, 0, len(in.Report))
	for i := range in.Report {
		returns = append(returns, Return{
			Barcode:          in.Report[i].Barcode,
			Brand:            in.Report[i].Brand,
			CompletedDt:      in.Report[i].CompletedDt,
			DstOfficeAddress: in.Report[i].DstOfficeAddress,
			DstOfficeId:      in.Report[i].DstOfficeId,
			ExpiredDt:        in.Report[i].ExpiredDt,
			IsStatusActive:   in.Report[i].IsStatusActive,
			NmId:             in.Report[i].NmId,
			OrderDt:          in.Report[i].OrderDt,
			OrderId:          in.Report[i].OrderId,
			ReadyToReturnDt:  in.Report[i].ReadyToReturnDt,
			Reason:           in.Report[i].Reason,
			ReturnType:       in.Report[i].ReturnType,
			ShkId:            in.Report[i].ShkId,
			Srid:             in.Report[i].Srid,
			Status:           in.Report[i].Status,
			StickerId:        in.Report[i].StickerId,
			SubjectName:      in.Report[i].SubjectName,
			TechSize:         in.Report[i].TechSize,
		})
	}

	return returns
}
