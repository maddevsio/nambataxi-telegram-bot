package states

import (
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
)

func NeedFare(service *holder.Service, session *storage.Session) {
	fareID, err := chat.GetFareIdByName(service.Update.Message.Text)
	if err != nil {
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ERROR_GET_1_FARE)
		msg.ReplyMarkup = chat.GetFaresKeyboard()
		service.Bot.Send(msg)
		return
	}
	session.FareId = fareID
	session.State = storage.STATE_NEED_ADDRESS
	service.DB.Save(&session)
	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ASK_ADDRESS)
	addresses := storage.GetLastAddressByChatID(service.DB, session.ChatID)
	if len(addresses) > 0 {
		msg.ReplyMarkup = chat.GetAddressKeyboard(addresses)
	}
	service.Bot.Send(msg)

}
