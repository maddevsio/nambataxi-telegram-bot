package handlers

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"fmt"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
)

func Fares(service *holder.Service) {
	fares, err := service.NambaTaxiAPI.GetFares()
	if err != nil {
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ERROR_GET_FARES)
		msg.ReplyMarkup = chat.GetBasicKeyboard()
		service.Bot.Send(msg)
		return
	}

	var faresText string
	for _, fare := range fares.Fare {
		faresText = faresText + fmt.Sprintf(chat.BOT_FARE_INFO,
			fare.Name,
			fare.Flagfall,
			fare.Cost_per_kilometer,
		)
	}

	faresText = faresText + chat.BOT_FARES_LINK

	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, faresText)
	msg.ReplyMarkup = chat.GetBasicKeyboard()
	msg.ParseMode = "Markdown"
	service.Bot.Send(msg)

}
