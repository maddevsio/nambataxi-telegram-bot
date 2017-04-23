package states

import (
	"strings"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
)

func NeedPhone(service *holder.Service, session *storage.Session, chatID int64) {
	phone := service.Update.Message.Text
	if service.Update.Message.Contact != nil {
		phone = "+" + service.Update.Message.Contact.PhoneNumber
	}

	if !strings.HasPrefix(phone, "+996") {
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_PHONE_START_996)
		phones := storage.GetLastPhonesByChatID(service.DB, chatID)
		if len(phones) > 0 {
			msg.ReplyMarkup = chat.GetPhonesKeyboard(phones)
		} else {
			msg.ReplyMarkup = chat.GetPhoneKeyboard()
		}
		service.Bot.Send(msg)
		return
	}
	session.Phone = phone
	session.State = storage.STATE_NEED_FARE
	service.DB.Save(&session)
	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ASK_FARE)
	msg.ReplyMarkup = chat.GetFaresKeyboard()
	service.Bot.Send(msg)

}
