package db

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

const (
	EnabledStatus = iota + 1
	DisabledStatus
	DeletedStatus
	WaitingWbState
	WaitingYaState
)

type Service struct {
	dsn string
}

func NewService(dsn string) Service {
	return Service{dsn: dsn}
}

func (dbs Service) InitDB() (*gorm.DB, error) {
	log.Println("Инициализация базы данных")
	Database, err := gorm.Open(postgres.Open(dbs.dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	return Database, nil
}

func (Stock) TableName() string {
	return "public.stocks"
}

func (User) TableName() string {
	return "public.users"
}

func (Cabinet) TableName() string {
	return "public.cabinets"
}

func (Order) TableName() string {
	return "public.orders"
}
