package db

import "time"

type Stock struct {
	ID          int       `gorm:"column:stock_id;unique"`
	Article     string    `gorm:"column:article"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	Marketplace string    `gorm:"column:marketplace"`
	StocksFBO   *int      `gorm:"column:stocks_fbo"`
	StocksFBS   *int      `gorm:"column:stocks_fbs"`
}

type User struct {
	ID     int   `gorm:"column:user_id;unique"`
	TgId   int64 `gorm:"column:tg_id;unique"`
	Status int   `gorm:"column:status"`
}
