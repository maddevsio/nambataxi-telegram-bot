package main

import (
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"github.com/maddevsio/nambataxi-telegram-bot/chat/handlers"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"github.com/maddevsio/nambataxi-telegram-bot/chat/handlers/states"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"

)

func chatStateMachine(service *holder.Service) {
	basicKeyboard := chat.GetBasicKeyboard()
	orderKeyboard := chat.GetOrderKeyboard()
	chatID := service.Update.Message.Chat.ID
	session := storage.GetSessionByChatID(service.DB, chatID)

	if session.ChatID != int64(0) {

		if service.Update.Message.Text == chat.BOT_MESSAGE_CANCEL_COMMAND {
			handlers.CancelNonCreatedOrder(service, chatID)
			return
		}

		if service.Update.Message.Text == chat.BOT_MESSAGE_NEAREST_CARS {
			handlers.NearestDrivers(service, &session)
			return
		}

		switch session.State {

		case storage.STATE_NEED_PHONE:
			states.NeedPhone(service, &session, chatID)
			return

		case storage.STATE_NEED_FARE:
			states.NeedFare(service, &session)
			return

		case storage.STATE_NEED_ADDRESS:
			chat.HandleOrderCreate(service, &session)
			chat.StartStatusReactionGoroutine(service, session)
			return

		case storage.STATE_ORDER_CREATED:
			if service.Update.Message.Text == chat.BOT_MESSAGE_CANCEL {
				chat.HandleOrderCancel(service, &session)
				return
			}

			order, err := service.NambaTaxiAPI.GetOrder(session.OrderId)
			if err != nil {
				msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, fmt.Sprintf(chat.BOT_ERROR_GET_ORDER, err))
				msg.ReplyMarkup = orderKeyboard
				service.Bot.Send(msg)
				return
			}

			chat.OrderStatusReaction(service, order, session)
			return

		default:
			service.DB.Delete(&session)
			msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ORDER_NOT_CREATED)
			msg.ReplyToMessageID = service.Update.Message.MessageID
			msg.ReplyMarkup = basicKeyboard
			service.Bot.Send(msg)
			return
		}
	}

	// messages reactions while out of session scope

	if service.Update.Message.Text == chat.BOT_MESSAGE_ORDER_FAST_START {
		handlers.OrderCreation(service)
		return
	}

	if service.Update.Message.Text == chat.BOT_MESSAGE_FARES {
		handlers.Fares(service)
		return
	}

	if service.Update.Message.Text == chat.BOT_MESSAGE_MY_ORDER_STATUS {
		handlers.OrderStatus(service)
		return
	}

	if service.Update.Message.Text == chat.BOT_MESSAGE_START_COMMAND {
		handlers.Start(service)
		return
	}

	handlers.Wuuut(service)
	return
}
