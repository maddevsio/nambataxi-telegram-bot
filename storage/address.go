package storage

import "github.com/jinzhu/gorm"

type Address struct {
	gorm.Model
	Text string
}

func GetLastAddressByChatID(db *gorm.DB, chatID int64) []Address {
	address := []Address{}
	db.Where("chat_id = ?", chatID).Find(&address).Limit(3)
	return address
}

