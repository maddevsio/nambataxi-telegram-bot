package storage

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndGetPhones(t *testing.T) {
	db := initDB()

	phone1 := Phone{}
	phone1.ChatID = 1
	phone1.Text = "+996111222333"
	db.Create(&phone1)

	phone2 := Phone{}
	phone2.ChatID = 1
	phone2.Text = "+996111222334"
	db.Create(&phone2)

	phone3 := Phone{}
	phone3.ChatID = 1
	phone3.Text = "+996111222335"
	db.Create(&phone3)

	phone4 := Phone{}
	phone4.ChatID = 1
	phone4.Text = "+996111222336"
	db.Create(&phone4)

	phones := GetLastPhonesByChatID(db, 1)
	assert.Len(t, phones, 3)
	assert.Equal(t, "+996111222336", phones[0].Text)

	deleteDB()
}

