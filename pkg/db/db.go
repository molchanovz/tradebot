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

type DataBaseService struct {
	dsn string
}

func NewDataBaseService(dsn string) DataBaseService {
	return DataBaseService{dsn: dsn}
}

func (dbs DataBaseService) InitDB() (*gorm.DB, error) {
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
