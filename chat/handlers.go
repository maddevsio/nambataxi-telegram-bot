package chat

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"log"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/jinzhu/gorm"
	"strconv"
	"fmt"
)

func HandleOrderCancel(nambaTaxiAPI api.NambaTaxiApi, session storage.Session, db *gorm.DB, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var message string
	var keyboard = GetOrderKeyboard()

	cancel, err := nambaTaxiAPI.CancelOrder(session.OrderId)
	if err != nil {
		message = "Произошла системная ошибка. Попробуйте еще раз"
		log.Printf("Error canceling order %v", err)
	}

	if cancel.Status == "200" {
		message = "Ваш заказ отменен"
		keyboard = GetBasicKeyboard()
		db.Delete(&session)
	}
	if cancel.Status == "400" {
		message = "Ваш заказ уже нельзя отменить, он передан водителю"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func HandleOrderCreate(nambaTaxiAPI api.NambaTaxiApi, session storage.Session, db *gorm.DB, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	session.Address = update.Message.Text
	orderOptions := map[string][]string{
		"phone_number": {session.Phone},
		"address":      {session.Address},
		"fare":         {strconv.Itoa(session.FareId)},
	}

	order, err := nambaTaxiAPI.MakeOrder(orderOptions)
	if err != nil {
		db.Delete(&session)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка создания заказа. Попробуйте еще раз")
		bot.Send(msg)
		return
	}
	session.State = storage.STATE_ORDER_CREATED
	session.OrderId = order.OrderId
	db.Save(&session)

	address := storage.Address{}
	address.ChatID = session.ChatID
	address.Text = session.Address
	db.FirstOrCreate(&address, storage.Address{ChatID: address.ChatID, Text: address.Text})

	phone := storage.Phone{}
	phone.ChatID = session.ChatID
	phone.Text = session.Phone
	db.FirstOrCreate(&phone, storage.Phone{ChatID: phone.ChatID, Text: phone.Text})

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Заказ создан! Номер заказа %v", order.OrderId))
	msg.ReplyMarkup = GetOrderKeyboard()
	bot.Send(msg)
}