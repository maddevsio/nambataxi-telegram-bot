package storage

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/jinzhu/gorm"
	"os"
)

const TEST_DB_NAME = "test.db"

func initDB() *gorm.DB {
	db := GetGormDB(TEST_DB_NAME)
	MigrateAll(db)
	return db
}

func deleteDB() {
	os.Remove(TEST_DB_NAME)
}

func getSession() Session {
	session := Session{}
	session.ChatID = 1
	session.OrderId = 1
	session.FareId = 1
	session.Address = "address"
	session.Phone = "123"
	session.State = "state"
	return session
}

func TestStoreAndGetSession(t *testing.T) {
	db := initDB()
	session1 := getSession()
	db.Create(&session1)
	session2 := Session{}
	db.First(&session2, "phone = ?", "123")
	assert.Equal(t, "123", session2.Phone)
	deleteDB()
}

func TestGetSessionByChatID_AndUpdate(t *testing.T) {
	db := initDB()
	session1 := getSession()
	db.Create(&session1)

	session2 := GetSessionByChatID(db, 1)
	session2.Phone = "456"
	db.Save(session2)

	session3 := GetSessionByChatID(db, 1)
	assert.Equal(t, "456", session3.Phone)
	deleteDB()
}

func TestGetSessionByChatID_DoesNotExist(t *testing.T) {
	db := initDB()
	session := GetSessionByChatID(db, 1)
	assert.Equal(t, Session{}, session)
	deleteDB()
}