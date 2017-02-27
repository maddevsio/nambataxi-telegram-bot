package chat

import (
	"testing"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestStoreSession(t *testing.T) {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Session{})

	session1 := Session{}
	session1.OrderId = 1
	session1.FareId = 1
	session1.Address = "address"
	session1.Phone = "123"
	session1.State = "state"

	// Create
	db.Create(&session1)

	session2 := Session{}
	db.First(&session2, "phone = ?", "123") // find product with id 1

	assert.Equal(t, "123", session2.Phone)
}