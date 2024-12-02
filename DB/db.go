package DB

import (
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var Database *gorm.DB

// Инициализация подключения
func InitDB() (*sql.DB, error) {
	dsn := "host=localhost user=postgres password=kvashok2002 dbname=db_bot port=5432 sslmode=disable"
	var err error
	Database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	sqlDB, err := Database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}
	return sqlDB, nil
}

type Stock struct {
	Article string    `gorm:"column:article;unique"`
	Date    time.Time `gorm:"column:date"`
	Stock   *int      `gorm:"column:stock"`
}

func (Stock) TableName() string {
	return "public.wb_stocks"
}
