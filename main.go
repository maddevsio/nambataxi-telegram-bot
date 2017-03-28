package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/maddevsio/simple-config"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	nambaTaxiAPI api.NambaTaxiApi

	config   = simple_config.NewSimpleConfig("config", "yml")
	db       = storage.GetGormDB(config.GetString("db_path"))
)

func main() {
	storage.MigrateAll(db)

	nambaTaxiAPI = api.NewNambaTaxiApi(
		config.GetString("partner_id"),
		config.GetString("server_token"),
		config.GetString("url"),
		config.GetString("version"),
	)

	chat.NambaTaxiApi = nambaTaxiAPI //init this for keyboards

	bot, err := tgbotapi.NewBotAPI(config.GetString("bot_token"))
	if err != nil {
		log.Panicf("Error connecting to Telegram: %v", err)
	}

	bot.Debug = config.Get("bot_debug").(bool)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Panicf("Error getting updates channel %v", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		chatStateMachine(update, bot)
	}
}

func chatStateMachine(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	basicKeyboard := chat.GetBasicKeyboard()
	orderKeyboard := chat.GetOrderKeyboard()
	session := storage.GetSessionByChatID(db, update.Message.Chat.ID)

	if session.ChatID != int64(0) {
		switch session.State {

		case storage.STATE_NEED_PHONE:
			phone := update.Message.Text
			if update.Message.Contact != nil {
				phone = "+" + update.Message.Contact.PhoneNumber
			}

			if !strings.HasPrefix(phone, "+996") {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Телефон должен начинаться с +996")
				phones := storage.GetLastPhonesByChatID(db, session.ChatID)
				if len(phones) > 0 {
					msg.ReplyMarkup = chat.GetPhonesKeyboard(phones)
				} else {
					msg.ReplyMarkup = chat.GetPhoneKeyboard()
				}
				bot.Send(msg)
				return
			}
			session.Phone = phone
			session.State = storage.STATE_NEED_FARE
			db.Save(&session)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Телефон сохранен. Теперь укажите тариф")
			msg.ReplyMarkup = chat.GetFaresKeyboard()
			bot.Send(msg)
			return

		case storage.STATE_NEED_FARE:
			fareID, err := chat.GetFareIdByName(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка! Не удалось получить тариф по имени. Попробуйте еще раз")
				msg.ReplyMarkup = chat.GetFaresKeyboard()
				bot.Send(msg)
				return
			}
			session.FareId = fareID
			session.State = storage.STATE_NEED_ADDRESS
			db.Save(&session)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Укажите ваш адрес. Куда подать машину?")
			addresses := storage.GetLastAddressByChatID(db, session.ChatID)
			if len(addresses) > 0 {
				msg.ReplyMarkup = chat.GetAddressKeyboard(addresses)
			}
			bot.Send(msg)
			return

		case storage.STATE_NEED_ADDRESS:
			// TODO: need to pass a structure, not this old-school list of params
			chat.HandleOrderCreate(nambaTaxiAPI, &session, db, update, bot)
			chat.StartStatusReactionGoroutine(nambaTaxiAPI, update, bot, db, session)
			return

		case storage.STATE_ORDER_CREATED:
			if update.Message.Text == "Отменить мой заказ" {
				chat.HandleOrderCancel(nambaTaxiAPI, &session, db, update, bot)
				return
			}

			order, err := nambaTaxiAPI.GetOrder(session.OrderId)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка получения заказа: %v", err))
				msg.ReplyMarkup = orderKeyboard
				bot.Send(msg)
				return
			}

			chat.OrderStatusReaction(order, update, bot, db, session)
			return

		default:
			db.Delete(&session)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заказ не открыт. Откройте заново")
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ReplyMarkup = basicKeyboard
			bot.Send(msg)
			return
		}
	}

	// messages reactions while out of session scope

	if update.Message.Text == "Быстрый заказ такси" {
		session := &storage.Session{}
		session.ChatID = update.Message.Chat.ID
		session.State = storage.STATE_NEED_PHONE
		db.Create(&session)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Укажите ваш телефон. Например: +996555112233")
		phones := storage.GetLastPhonesByChatID(db, session.ChatID)
		if len(phones) > 0 {
			msg.ReplyMarkup = chat.GetPhonesKeyboard(phones)
		} else {
			msg.ReplyMarkup = chat.GetPhoneKeyboard()
		}
		bot.Send(msg)
		return
	}

	if update.Message.Text == "Тарифы" {
		fares, err := nambaTaxiAPI.GetFares()
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка. Не удалось получить тарифы. Попробуйте еще раз")
			msg.ReplyMarkup = basicKeyboard
			bot.Send(msg)
			return
		}

		var faresText string
		for _, fare := range fares.Fare {
			faresText = faresText + fmt.Sprintf("Тариф: %v. Стоимость посадки: %.2f. Стоимость за километр: %.2f.\n\n",
				fare.Name,
				fare.Flagfall,
				fare.Cost_per_kilometer,
			)
		}

		faresText = faresText + "Для получения подробной информации посетите https://nambataxi.kg/ru/tariffs/"

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, faresText)
		msg.ReplyMarkup = basicKeyboard
		bot.Send(msg)
		return
	}

	if update.Message.Text == "Узнать статус моего заказа" {
		session := storage.GetSessionByChatID(db, update.Message.Chat.ID)
		db.Delete(&session)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "К сожалению у вас нет заказа")
		msg.ReplyMarkup = basicKeyboard
		bot.Send(msg)
		return
	}

	if update.Message.Text == "/start" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вас приветствует бот Намба Такси для мессенджера Телеграм")
		msg.ReplyMarkup = basicKeyboard
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Что-что?")
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = basicKeyboard
	bot.Send(msg)
	return
}
