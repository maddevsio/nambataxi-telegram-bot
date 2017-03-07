package storage

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndGetAddress(t *testing.T) {
	db := initDB()

	address1 := Address{}
	address1.ChatID = 1
	address1.Text = "Address 1"
	db.Create(&address1)

	address2 := Address{}
	address2.ChatID = 1
	address2.Text = "Address 2"
	db.Create(&address2)

	address3 := Address{}
	address3.ChatID = 1
	address3.Text = "Address 3"
	db.Create(&address3)

	address4 := Address{}
	address4.ChatID = 1
	address4.Text = "Address 4"
	db.Create(&address4)

	address := GetLastAddressByChatID(db, 1)
	assert.Len(t, address, 3)
	assert.Equal(t, "Address 4", address[0].Text)

	deleteDB()
}

