package chat

import (
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"strconv"
	"fmt"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
)

func HandleOrderCancel(service *holder.Service, session *storage.Session) {
	var message string
	var keyboard = GetOrderKeyboard()

	cancel, err := service.NambaTaxiAPI.CancelOrder(session.OrderId)
	if err != nil {
		message = "Произошла системная ошибка. Попробуйте еще раз"
		log.Printf("Error canceling order %v", err)
	}

	if cancel.Status == "200" {
		message = BOT_ORDER_CANCELED_BY_USER
		keyboard = GetBasicKeyboard()
		service.DB.Delete(session)
	}
	if cancel.Status == "400" {
		message = "Ваш заказ уже нельзя отменить, он передан водителю"
	}

	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, message)
	msg.ReplyMarkup = keyboard
	service.Bot.Send(msg)
}

func HandleOrderCreate(service *holder.Service, session *storage.Session) {
	session.Address = service.Update.Message.Text
	orderOptions := map[string][]string{
		"phone_number": {session.Phone},
		"address":      {session.Address},
		"fare":         {strconv.Itoa(session.FareId)},
	}

	order, err := service.NambaTaxiAPI.MakeOrder(orderOptions)
	if err != nil {
		service.DB.Delete(session)
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, "Ошибка создания заказа. Попробуйте еще раз")
		service.Bot.Send(msg)
		return
	}
	session.State = storage.STATE_ORDER_CREATED
	session.OrderId = order.OrderId
	service.DB.Save(session)

	address := storage.Address{}
	address.ChatID = session.ChatID
	address.Text = session.Address
	service.DB.FirstOrCreate(&address, storage.Address{ChatID: address.ChatID, Text: address.Text})

	phone := storage.Phone{}
	phone.ChatID = session.ChatID
	phone.Text = session.Phone
	service.DB.FirstOrCreate(&phone, storage.Phone{ChatID: phone.ChatID, Text: phone.Text})

	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf(BOT_ORDER_CREATED, order.OrderId))
	msg.ReplyMarkup = GetOrderKeyboard()
	service.Bot.Send(msg)
}