package handlers

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
)

func OrderCreation(service *holder.Service) {
	session := &storage.Session{}
	session.ChatID = service.Update.Message.Chat.ID
	session.State = storage.STATE_NEED_PHONE
	service.DB.Create(&session)
	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ASK_PHONE)
	phones := storage.GetLastPhonesByChatID(service.DB, service.Update.Message.Chat.ID)
	if len(phones) > 0 {
		msg.ReplyMarkup = chat.GetPhonesKeyboard(phones)
	} else {
		msg.ReplyMarkup = chat.GetPhoneKeyboard()
	}
	service.Bot.Send(msg)

}
