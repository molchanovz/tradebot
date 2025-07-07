package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testRepo *Repo

func TestMain(m *testing.M) {
	testRepo, _ = NewRepo("postgres://sergey:1719@localhost:5432/tradebot?sslmode=disable")
	m.Run()
}

func TestGetCabinets(t *testing.T) {
	cabinets, err := testRepo.GetCabinets("OZON")
	assert.Nil(t, err)
	t.Log(cabinets)

}

func TestGetCabinetById(t *testing.T) {
	got, err := testRepo.GetCabinetById("1")
	assert.Nil(t, err)
	t.Log(got)

}

func TestRepo_GetUserByTgId(t *testing.T) {
	got, err := testRepo.GetUserByTgId(406363099)
	assert.Nil(t, err)
	t.Log(got)
}
