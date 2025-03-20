package db

import "time"

type Stock struct {
	Article   string    `gorm:"column:article;unique"`
	Date      time.Time `gorm:"column:date"`
	StocksFBO *int      `gorm:"column:stock_fbo"`
	StocksFBS *int      `gorm:"column:stock_fbs"`
}

type User struct {
	ChatId int64 `gorm:"column:chatid;unique"`
	State  int   `gorm:"column:state"`
}
