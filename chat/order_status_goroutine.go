package chat

import (
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"time"
	"log"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
)

func StartStatusReactionGoroutine(service *holder.Service, session storage.Session) {
	go func(session storage.Session) {
		var status = "Новый заказ"
		for {
			time.Sleep(5 * time.Second)
			log.Printf("Session order id: %v", session.OrderId)
			currentOrder, err := service.NambaTaxiAPI.GetOrder(session.OrderId)
			if err != nil {
				log.Printf("Error getting order status %v", err)
				return
			} else {
				log.Printf("Order status %v", currentOrder.Status)
			}
			if status != currentOrder.Status {
				if currentOrder.Status == "Отклонен" {
					fetchedSession := storage.GetSessionByChatID(service.DB, session.ChatID)
					if fetchedSession.ChatID == 0 {
						// This means that the user has canceled the order already
						// and we do not need to send him a message from this goroutine.
						// When a operator cancels the message, this "if" will be not active
						// and user will get a message
						return
					}
				}
				OrderStatusReaction(service, currentOrder, session)
			}
			status = currentOrder.Status
			if currentOrder.Status == "Отклонен" || currentOrder.Status == "Выполнен" {
				return
			}
		}
	}(session)
}

func OrderStatusReaction(service *holder.Service, order api.Order, session storage.Session) {
	basicKeyboard := GetBasicKeyboard()
	orderKeyboard := GetOrderKeyboard()

	var msg tgbotapi.MessageConfig

	if order.Status == "Новый заказ" {
		msg = tgbotapi.NewMessage(service.Update.Message.Chat.ID, BOT_ORDER_THANKS)
		msg.ReplyMarkup = orderKeyboard
		service.Bot.Send(msg)
		return
	}

	if order.Status == "Принят" {
		msg = tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf(
			BOT_ORDER_ACCEPTED,
			order.Driver.CabNumber,
			order.Driver.Name,
			order.Driver.PhoneNumber,
			order.Driver.LicensePlate,
			order.Driver.Make,
		))
		msg.ReplyMarkup = orderKeyboard
		service.Bot.Send(msg)

		if order.Driver.Lat == float64(0) || order.Driver.Lon == float64(0) {
			log.Print("Driver with empty lat or long")
			return
		}

		msg = tgbotapi.NewMessage(service.Update.Message.Chat.ID, BOT_DRIVER_LOCATION)
		msg.ReplyMarkup = orderKeyboard
		service.Bot.Send(msg)
		loc := tgbotapi.NewLocation(service.Update.Message.Chat.ID, order.Driver.Lat, order.Driver.Lon)
		loc.ReplyMarkup = orderKeyboard
		service.Bot.Send(loc)


		return
	}

	if order.Status == "Выполнен" {
		service.DB.Delete(&session)
		msg = tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf(BOT_ORDER_DONE, order.TripCost))
		msg.ReplyMarkup = basicKeyboard
		msg.ParseMode = "Markdown"
		service.Bot.Send(msg)
		return
	}

	if order.Status == "Отклонен" {
		service.DB.Delete(&session)
		msg = tgbotapi.NewMessage(service.Update.Message.Chat.ID, BOT_ORDER_CANCELED_BY_OPERATOR)
		msg.ReplyMarkup = basicKeyboard
		service.Bot.Send(msg)
		return
	}

	msg = tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf("%v", order.Status))
	msg.ReplyMarkup = orderKeyboard
	service.Bot.Send(msg)
	return
}
