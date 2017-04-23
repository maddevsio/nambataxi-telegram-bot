package handlers

import (
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
)

func NearestDrivers(service *holder.Service, session *storage.Session) {
	if session.State == storage.STATE_ORDER_CREATED {
		nearestDrivers, err := service.NambaTaxiAPI.GetNearestDrivers(session.Address)
		if err != nil {
			msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf(chat.BOT_ERROR_GET_NEAREST_DRIVERS, err))
			msg.ReplyMarkup = chat.GetOrderKeyboard()
			service.Bot.Send(msg)
			return
		}
		if nearestDrivers.Status != "200" {
			msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf(chat.BOT_ERROR_GET_NEAREST_DRIVERS, nearestDrivers.Message))
			msg.ReplyMarkup = chat.GetOrderKeyboard()
			service.Bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf(chat.BOT_NEAREST_DRIVERS, nearestDrivers.Drivers))
		msg.ReplyMarkup = chat.GetOrderKeyboard()
		service.Bot.Send(msg)
		return
	} else {
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ERROR_EARLY_NEAREST_DRIVERS)
		service.Bot.Send(msg)
		return
	}
}