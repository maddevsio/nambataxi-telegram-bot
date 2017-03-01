package storage

import (
	"github.com/jinzhu/gorm"
)

func GetGormDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func MigrateAll(db *gorm.DB ) {
	db.AutoMigrate(&Session{})
}