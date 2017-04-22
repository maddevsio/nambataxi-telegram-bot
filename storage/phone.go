package storage

import (
	"github.com/jinzhu/gorm"
)

type Phone struct {
	gorm.Model
	ChatID int64
	Text string
}

func GetLastPhonesByChatID(db *gorm.DB, chatID int64) []Phone {
	phones := []Phone{}
	db.Order("id desc").Limit(3).Where("chat_id = ?", chatID).Find(&phones)
	return phones
}

