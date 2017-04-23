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
			states.NeedFare(service, &session)
			return

		case storage.STATE_NEED_ADDRESS:
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
		handlers.OrderCreation(service)
		return
	}

	if service.Update.Message.Text == "Тарифы" {
		handlers.Fares(service)
		return
	}

	if service.Update.Message.Text == "Узнать статус моего заказа" {
		handlers.OrderStatus(service)
		return
	}

	if service.Update.Message.Text == "/start" {
		handlers.Start(service)
		return
	}

	handlers.Wuuut(service)
	return
}
