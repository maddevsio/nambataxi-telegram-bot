package storage

import (
	"testing"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/jinzhu/gorm"
)

func initDB() *gorm.DB {
	db := GetGormDB()
	MigrateAll(db)
	return db
}

func TestStoreSession(t *testing.T) {
	db := initDB()

	session1 := Session{}
	session1.ChatID = 1
	session1.OrderId = 1
	session1.FareId = 1
	session1.Address = "address"
	session1.Phone = "123"
	session1.State = "state"

	db.Create(&session1)

	session2 := Session{}
	db.First(&session2, "phone = ?", "123") // find product with id 1

	assert.Equal(t, "123", session2.Phone)
}