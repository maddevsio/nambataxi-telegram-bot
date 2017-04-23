package handlers

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
)

func Wuuut(service *holder.Service) {
	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, "Что-что?")
	msg.ReplyToMessageID = service.Update.Message.MessageID
	msg.ReplyMarkup = chat.GetBasicKeyboard()
	service.Bot.Send(msg)
}
