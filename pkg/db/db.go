package db

import (
	"errors"
	"github.com/go-pg/pg/v10"
	"log"
)

const (
	EnabledStatus = iota + 1
	DisabledStatus
	DeletedStatus
	WaitingWbState
	WaitingYaState
)

type Repo struct {
	DB *pg.DB
}

func NewRepo(dsn string) (*Repo, error) {
	dbc, err := initDB(dsn)
	if err != nil {
		return nil, err
	}
	return &Repo{DB: dbc}, nil
}

func initDB(dsn string) (*pg.DB, error) {
	log.Println("Инициализация базы данных")
	options, err := pg.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	dbc := pg.Connect(options)
	return dbc, nil
}

func (r Repo) GetCabinets(marketplace string) ([]Cabinet, error) {
	var cabinets []Cabinet
	err := r.DB.Model(&cabinets).Where(`"marketplace" = ?`, marketplace).Select()
	return cabinets, err
}

func (r Repo) GetCabinetById(id string) (Cabinet, error) {
	var cabinet Cabinet
	err := r.DB.Model(&cabinet).Where(`"cabinetsId" = ?`, id).Select()
	return cabinet, err
}

func (r Repo) GetUserByTgId(tgId int64) (User, error) {
	var user User
	err := r.DB.Model(&user).Where(`"tgId" = ?`, tgId).Select()
	if errors.Is(err, pg.ErrNoRows) {
		return user, nil
	} else if err != nil {
		return user, err
	}

	return user, nil
}

func (r Repo) GetPrintedOrders(marketplace string) ([]Order, error) {
	var printedOrders []Order
	err := r.DB.Model(printedOrders).Where(`"marketplace" = ?`, marketplace).Select()
	return printedOrders, err
}

func (r Repo) CreateOrders(orders []Order) error {
	_, err := r.DB.Model(orders).Insert()
	return err
}

func (r Repo) UpdateUser(u User) error {
	_, err := r.DB.Model(&u).Where(`"tgId" = ?`, u.TgID).Update()
	return err
}

func (r Repo) CreateUser(u User) error {
	_, err := r.DB.Model(&u).Insert()
	return err
}

func (r Repo) GetStocks(article, cabinetId string) ([]Stock, error) {
	var stocks []Stock
	err := r.DB.Model(stocks).Where("article = ? and cabinetId = ?", article, cabinetId).Select()
	return stocks, err
}

func (r Repo) CreateStock(s Stock) error {
	_, err := r.DB.Model(&s).Insert()
	return err
}

func (r Repo) UpdateStock(stock Stock) error {
	_, err := r.DB.Model(&stock).Where("article = ? and cabinetId = ?", stock.Article, stock.CabinetID).Update()
	return err
}
