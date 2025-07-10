package marketplaces

import "tradebot/pkg/db"

type Authorization struct {
	ClientId, Token, Type string
}

type Cabinet struct {
	ID       int
	Name     string
	ClientId string
	Key      string
}

type Cabinets []Cabinet

func NewCabinet(in db.Cabinet) Cabinet {
	return Cabinet{
		ID:       in.ID,
		Name:     in.Name,
		ClientId: *in.ClientID,
		Key:      in.Key,
	}
}

func NewCabinets(in []db.Cabinet) Cabinets {
	newCabinets := Cabinets{}
	for _, c := range in {
		newCabinets = append(newCabinets, NewCabinet(c))
	}
	return newCabinets
}
