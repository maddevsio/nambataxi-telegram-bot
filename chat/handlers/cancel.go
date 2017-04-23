package handlers

import (
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
)

func CancelNonCreatedOrder(service *holder.Service, chatID int64) {
	session := storage.GetSessionByChatID(service.DB, chatID)
	chat.HandleOrderCancel(service, &session)
	service.DB.Delete(&session)
	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_WELCOME_MESSAGE)
	msg.ReplyMarkup = chat.GetBasicKeyboard()
	service.Bot.Send(msg)
}