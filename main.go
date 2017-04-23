package main

import (
	"fmt"
	"log"

	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/maddevsio/simple-config"
	"gopkg.in/telegram-bot-api.v4"
	"strconv"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"github.com/maddevsio/nambataxi-telegram-bot/chat/handlers"
	"github.com/maddevsio/nambataxi-telegram-bot/chat/handlers/states"
)

var (
	service  holder.Service
	err error
)

func main() {
	service.Config = simple_config.NewSimpleConfig("config", "yml")
	service.DB = storage.GetGormDB(service.Config.GetString("db_path"))

	storage.MigrateAll(service.DB)

	service.NambaTaxiAPI = api.NewNambaTaxiApi(
		service.Config.GetString("partner_id"),
		service.Config.GetString("server_token"),
		service.Config.GetString("url"),
		service.Config.GetString("version"),
	)

	chat.NambaTaxiApi = service.NambaTaxiAPI //init this for keyboards

	service.Bot, err = tgbotapi.NewBotAPI(service.Config.GetString("bot_token"))
	if err != nil {
		log.Panicf("Error connecting to Telegram: %v", err)
	}

	service.Bot.Debug, err = strconv.ParseBool(service.Config.GetString("bot_debug"))
	if err != nil {
		log.Panicf("Cannot convert debug status from config: %v", err)
	}

	log.Printf("Authorized on account %s", service.Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := service.Bot.GetUpdatesChan(u)

	if err != nil {
		log.Panicf("Error getting updates channel %v", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		service.Update = update
		chatStateMachine(&service)
	}
}

func chatStateMachine(service *holder.Service) {
	basicKeyboard := chat.GetBasicKeyboard()
	orderKeyboard := chat.GetOrderKeyboard()
	chatID := service.Update.Message.Chat.ID
	session := storage.GetSessionByChatID(service.DB, chatID)

	if session.ChatID != int64(0) {

		if service.Update.Message.Text == "/Cancel" {
			handlers.CancelNonCreatedOrder(service, chatID)
			return
		}

		if service.Update.Message.Text == "Машины рядом" {
			handlers.NearestDrivers(service, &session)
			return
		}

		switch session.State {

		case storage.STATE_NEED_PHONE:
			states.NeedPhone(service, &session, chatID)
			return

		case storage.STATE_NEED_FARE:
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
			addresses := storage.GetLastAddressByChatID(service.DB, chatID)
			if len(addresses) > 0 {
				msg.ReplyMarkup = chat.GetAddressKeyboard(addresses)
			}
			service.Bot.Send(msg)
			return

		case storage.STATE_NEED_ADDRESS:
			// TODO: need to pass a structure, not this old-school list of params
			chat.HandleOrderCreate(service, &session)
			chat.StartStatusReactionGoroutine(service, session)
			return

		case storage.STATE_ORDER_CREATED:
			if service.Update.Message.Text == "Отменить мой заказ" {
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

	if service.Update.Message.Text == "Быстрый заказ такси" {
		session := &storage.Session{}
		session.ChatID = service.Update.Message.Chat.ID
		session.State = storage.STATE_NEED_PHONE
		service.DB.Create(&session)
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ASK_PHONE)
		phones := storage.GetLastPhonesByChatID(service.DB, chatID)
		if len(phones) > 0 {
			msg.ReplyMarkup = chat.GetPhonesKeyboard(phones)
		} else {
			msg.ReplyMarkup = chat.GetPhoneKeyboard()
		}
		service.Bot.Send(msg)
		return
	}

	if service.Update.Message.Text == "Тарифы" {
		fares, err := service.NambaTaxiAPI.GetFares()
		if err != nil {
			msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_ERROR_GET_FARES)
			msg.ReplyMarkup = basicKeyboard
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
		msg.ReplyMarkup = basicKeyboard
		msg.ParseMode = "Markdown"
		service.Bot.Send(msg)
		return
	}

	if service.Update.Message.Text == "Узнать статус моего заказа" {
		session := storage.GetSessionByChatID(service.DB, chatID)
		service.DB.Delete(&session)
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_NO_ORDERS)
		msg.ReplyMarkup = basicKeyboard
		service.Bot.Send(msg)
		return
	}

	if service.Update.Message.Text == "/start" {
		msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, chat.BOT_WELCOME_MESSAGE)
		msg.ReplyMarkup = basicKeyboard
		service.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(service.Update.Message.Chat.ID, "Что-что?")
	msg.ReplyToMessageID = service.Update.Message.MessageID
	msg.ReplyMarkup = basicKeyboard
	service.Bot.Send(msg)
	return
}
