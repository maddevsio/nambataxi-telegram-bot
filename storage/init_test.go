package storage

import (
	"github.com/jinzhu/gorm"
	"os"
)

func initDB() *gorm.DB {
	db := GetGormDB(TEST_DB_NAME)
	MigrateAll(db)
	return db
}

func deleteDB() {
	os.Remove(TEST_DB_NAME)
}

