package handlers

import (
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
)

func OrderStatus(service *holder.Service) {
	session := storage.GetSessionByChatID(service.DB, service.Update.Message.Chat.ID)
	service.DB.Delete(&session)
	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_NO_ORDERS)
	msg.ReplyMarkup = chat.GetBasicKeyboard()
	service.Bot.Send(msg)
}
