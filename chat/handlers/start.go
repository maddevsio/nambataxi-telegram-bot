package handlers

import (
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
)

func Start(service *holder.Service) {
	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_WELCOME_MESSAGE)
	msg.ReplyMarkup = chat.GetBasicKeyboard()
	service.Bot.Send(msg)
}
