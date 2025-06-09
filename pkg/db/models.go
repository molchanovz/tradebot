package db

import "time"

type Stock struct {
	ID          int       `gorm:"column:stockId;unique"`
	Article     string    `gorm:"column:article"`
	UpdatedAt   time.Time `gorm:"column:updatedAt"`
	Marketplace string    `gorm:"column:marketplace"`
	StocksFBO   *int      `gorm:"column:stocksFbo"`
	StocksFBS   *int      `gorm:"column:stocksFbs"`
}

type Order struct {
	ID            int    `gorm:"column:orderId;unique"`
	PostingNumber string `gorm:"column:postingNumber"`
	Marketplace   string `gorm:"column:marketplace"`
}

type User struct {
	ID       int   `gorm:"column:userId;unique"`
	TgId     int64 `gorm:"column:tgId;unique"`
	StatusId int   `gorm:"column:statusId"`
	IsAdmin  bool  `gorm:"column:isAdmin"`
}

type Cabinet struct {
	ID          int    `gorm:"column:cabinetsId;unique"`
	Name        string `gorm:"column:name"`
	ClientId    string `gorm:"column:clientId;unique"`
	Key         string `gorm:"column:key;unique"`
	Marketplace string `gorm:"column:marketplace"`
	Type        string `gorm:"column:type"`
	UserId      int    `gorm:"column:userId"`
}
