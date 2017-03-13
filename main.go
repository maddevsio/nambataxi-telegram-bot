package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"time"

	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/maddevsio/simple-config"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	nambaTaxiAPI api.NambaTaxiApi

	config   = simple_config.NewSimpleConfig("config", "yml")
	sessions = storage.GetAllSessions()
	db       = storage.GetGormDB("namba-taxi-bot.db")
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
			if !strings.HasPrefix(update.Message.Text, "+996") {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Телефон должен начинаться с +996")
				bot.Send(msg)
				return
			}
			session.Phone = update.Message.Text
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
				msg.ReplyMarkup = basicKeyboard
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
			session.Address = update.Message.Text
			orderOptions := map[string][]string{
				"phone_number": {session.Phone},
				"address":      {session.Address},
				"fare":         {strconv.Itoa(session.FareId)},
			}

			order, err := nambaTaxiAPI.MakeOrder(orderOptions)
			if err != nil {
				delete(sessions, update.Message.Chat.ID)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка создания заказа. Попробуйте еще раз")
				bot.Send(msg)
				return
			}
			session.State = storage.STATE_ORDER_CREATED
			session.OrderId = order.OrderId
			db.Save(&session)

			address := storage.Address{}
			address.ChatID = session.ChatID
			address.Text = session.Address
			db.FirstOrCreate(&address, storage.Address{ChatID: address.ChatID, Text: address.Text})

			phone := storage.Phone{}
			phone.ChatID = session.ChatID
			phone.Text = session.Phone
			db.FirstOrCreate(&phone, storage.Phone{ChatID: phone.ChatID, Text: phone.Text})

			go func() {
				var status = "Новый заказ"
				for {
					time.Sleep(5 * time.Second)
					log.Printf("Session order id: %v", session.OrderId)
					currentOrder, err := nambaTaxiAPI.GetOrder(session.OrderId)
					if err != nil {
						log.Printf("Error getting order status %v", err)
					} else {
						log.Printf("Order status %v", currentOrder.Status)
					}
					if status != currentOrder.Status {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%v", currentOrder.Status))
						msg.ReplyMarkup = chat.GetOrderKeyboard()
						bot.Send(msg)
					}
					status = currentOrder.Status
					if currentOrder.Status == "Отклонен" {
						return
					}
				}
			}()

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Заказ создан! Номер заказа %v", order.OrderId))
			msg.ReplyMarkup = chat.GetOrderKeyboard()
			bot.Send(msg)
			return

		case storage.STATE_ORDER_CREATED:
			if update.Message.Text == "Отменить мой заказ" {
				var message string
				var keyboard = orderKeyboard

				cancel, err := nambaTaxiAPI.CancelOrder(session.OrderId)
				if err != nil {
					message = "Произошла системная ошибка. Попробуйте еще раз"
					log.Printf("Error canceling order %v", err)
				}

				if cancel.Status == "200" {
					message = "Ваш заказ отменен"
					keyboard = basicKeyboard
					db.Delete(&session)
				}
				if cancel.Status == "400" {
					message = "Ваш заказ уже нельзя отменить, он передан водителю"
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
				return
			}

			order, err := nambaTaxiAPI.GetOrder(session.OrderId)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка получения заказа: %v", err))
				msg.ReplyMarkup = orderKeyboard
				bot.Send(msg)
				return
			}

			var msg tgbotapi.MessageConfig

			if order.Status == "Новый заказ" {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint("Спасибо за ваш заказ. Он находится в обработке. Мы нашли рядом с вами 3 свободных машины. Совсем скоро водитель возьмет ваш заказ"))
				msg.ReplyMarkup = orderKeyboard
				bot.Send(msg)
				return
			}

			if order.Status == "Принят" {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
					"Ура! Ваш заказ принят ближайшим водителем!\nНомер борта: %v\nВодитель: %v\nТелефон: %v\nГосномер: %v\nМарка машины: %v",
					order.Driver.CabNumber,
					order.Driver.Name,
					order.Driver.PhoneNumber,
					order.Driver.LicensePlate,
					order.Driver.Make,
				))
				msg.ReplyMarkup = orderKeyboard
				bot.Send(msg)
				return
			}

			if order.Status == "Выполнен" {
				db.Delete(&session)
				msg = tgbotapi.NewMessage(update.Message.Chat.ID,
					"Ваш заказ выполнен. Спасибо, что воспользовались услугами Намба Такси. Если вдруг что-то не так, то телефон Отдела Контроля Качества к вашим услугам:\n"+
						"+996 (312) 97-90-60\n"+
						"+996 (701) 97-67-03\n"+
						"+996 (550) 97-60-23",
				)
				msg.ReplyMarkup = basicKeyboard
				bot.Send(msg)
				return
			}

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%v", order.Status))
			msg.ReplyMarkup = orderKeyboard
			bot.Send(msg)
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
