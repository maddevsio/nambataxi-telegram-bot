package storage

import (
	"github.com/jinzhu/gorm"
)

type Address struct {
	gorm.Model
	ChatID int64
	Text string
}

func GetLastAddressByChatID(db *gorm.DB, chatID int64) []Address {
	address := []Address{}
	db.Order("id desc").Limit(3).Where("chat_id = ?", chatID).Find(&address)
	return address
}

