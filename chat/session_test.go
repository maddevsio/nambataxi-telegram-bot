package chat

import (
	"testing"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestStoreSession(t *testing.T) {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Session{})

	session := Session{}
	session.OrderId = 1
	session.FareId = 1
	session.Address = "address"
	session.Phone = "123"
	session.State = "state"

	// Create
	db.Create(&session)
}