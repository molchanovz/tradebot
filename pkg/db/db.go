package db

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

const (
	WaitingWbState = 1
	WaitingYaState = 2
	DefaultState   = 3
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
	return "public.wb_stocks"
}

func (User) TableName() string {
	return "public.users"
}
