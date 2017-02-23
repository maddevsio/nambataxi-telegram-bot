package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/simple-config"
	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"fmt"
	"strings"
	"errors"
	"strconv"
)

type Session struct {
	Phone string
	Address string
	FareId int
	State string
	Order api.Order
}

var (
	config = simple_config.NewSimpleConfig("config", "yml")
	sessions = make(map[int64]*Session)
	nambaTaxiApi api.NambaTaxiApi
)

const (
	STATE_NEED_PHONE    = "need phone"
	STATE_NEED_ADDRESS  = "need address"
	STATE_NEED_FARE     = "need fare"
	STATE_ORDER_CREATED = "order created"
)

func main() {
	nambaTaxiApi = api.NewNambaTaxiApi(
		config.GetString("partner_id"),
		config.GetString("server_token"),
		config.GetString("url"),
		config.GetString("version"),
	)

	bot, err := tgbotapi.NewBotAPI(config.GetString("bot_token"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		chatStateMachine(update, bot)
	}
}

func chatStateMachine (update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	keyboard := getBasicKeyboard()
	orderKeyboard := getOrderKeyboard()

	if session := sessions[update.Message.Chat.ID]; session != nil {
		switch session.State {

		case STATE_NEED_PHONE:
			if !strings.HasPrefix(update.Message.Text, "+996") {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Телефон должен начинаться с +996")
				bot.Send(msg)
				return
			}
			session.Phone = update.Message.Text
			session.State = STATE_NEED_FARE
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Телефон сохранен. Теперь укажите тариф")
			msg.ReplyMarkup = getFaresKeyboard()
			bot.Send(msg)
			return

		case STATE_NEED_FARE:
			fareId, err := getFareIdByName(update.Message.Text)
			if (err != nil) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка! Не удалось получить тариф по имени. Попробуйте еще раз")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
				return
			}
			session.FareId = fareId
			session.State = STATE_NEED_ADDRESS
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Укажите ваш адрес. Куда подать машину?")
			bot.Send(msg)
			return

		case STATE_NEED_ADDRESS:
			session.Address = update.Message.Text
			orderOptions := map[string][]string{
				"phone_number": {session.Phone},
				"address":      {session.Address},
				"fare":         {strconv.Itoa(session.FareId)},
			}

			order, err := nambaTaxiApi.MakeOrder(orderOptions)
			if err != nil {
				delete(sessions, update.Message.Chat.ID)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка создания заказа. Попробуйте еще раз")
				bot.Send(msg)
				return
			}
			session.State = STATE_ORDER_CREATED
			session.Order = order
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Заказ создан! Номер заказа %v", order.OrderId))
			msg.ReplyMarkup = getOrderKeyboard()
			bot.Send(msg)
			return

		case STATE_ORDER_CREATED:
			if update.Message.Text == "Отменить мой заказ" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мы пока не умеем отменять заказ. Извините.")
				msg.ReplyMarkup = orderKeyboard
				bot.Send(msg)
				return
			}

			order, err := nambaTaxiApi.GetOrder(session.Order.OrderId)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка получения заказа: %v", err))
				msg.ReplyMarkup = orderKeyboard
				bot.Send(msg)
				return
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Машина скоро будет. Статус вашего заказа: %v", order.Status))
				msg.ReplyMarkup = orderKeyboard
				bot.Send(msg)
				return
			}

		default:
			delete(sessions, update.Message.Chat.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Заказ не открыт. Откройте заново")
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			return
		}
	}

	if update.Message.Text == "Быстрый заказ такси" {
		sessions[update.Message.Chat.ID] = &Session{}
		sessions[update.Message.Chat.ID].State = STATE_NEED_PHONE
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Укажите ваш телефон. Например: +996555112233")
		bot.Send(msg)
		return
	}

	if update.Message.Text == "Узнать статус моего заказа" {
		delete(sessions, update.Message.Chat.ID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "К сожалению у вас нет заказа")
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Что-что?")
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
	return
}

func getBasicKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Быстрый заказ такси"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Тарифы"),
		),
	)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func getOrderKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Узнать статус моего заказа"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Отменить мой заказ"),
		),
	)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func getFaresKeyboard() tgbotapi.ReplyKeyboardMarkup {
	fares, err := nambaTaxiApi.GetFares()
	if err != nil {
		log.Printf("error getting fares: %v", err)
		return tgbotapi.NewReplyKeyboard()
	}

	var rows []tgbotapi.KeyboardButton
	for _, fare := range fares.Fare {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fare.Name))...)
	}

	keyboard := tgbotapi.NewReplyKeyboard(rows)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func getFareIdByName(fareName string) (int, error) {
	fares, err := nambaTaxiApi.GetFares()
	if err != nil {
		log.Printf("error getting fares: %v", err)
		return 0, err
	}
	for _, fare := range fares.Fare {
		if fare.Name == fareName {
			return fare.Id, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Cannot find fare with name %v", fareName))
}