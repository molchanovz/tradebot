package ozon

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"tradebot/pkg/db"
)

var testRepo *db.Repo
var cabinet db.Cabinet
var err error

func TestMain(m *testing.M) {
	testRepo, err = db.NewRepo("postgres://sergey:1719@localhost:5432/tradebot?sslmode=disable")
	if err != nil {
		return
	}
	cabinet, err = testRepo.GetCabinetById("3")
	if err != nil {
		return
	}
	m.Run()
}

func TestReturnsList(t *testing.T) {
	since := time.Now().AddDate(0, 0, -2).Format("2006-01-02") + "T21:00:00.000Z"
	to := time.Now().AddDate(0, 0, -1).Format("2006-01-02") + "T21:00:00.000Z"
	got, err := ReturnsList(*cabinet.ClientID, cabinet.Key, 0, since, to)
	assert.Nil(t, err)
	t.Log(got)
}
