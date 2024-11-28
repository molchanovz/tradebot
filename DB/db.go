package DB

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var Database *gorm.DB

// Инициализация подключения
func InitDB() error {
	dsn := "host=localhost user=postgres password=Kvashok2002 dbname=db_bot port=5432 sslmode=disable"
	var err error
	Database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	return nil
}

type Stock struct {
	Article string    `gorm:"column:article;unique"`
	Date    time.Time `gorm:"column:date"`
	Stock   *int      `gorm:"column:stock"`
}

func (Stock) TableName() string {
	return "public.wb_stocks"
}
