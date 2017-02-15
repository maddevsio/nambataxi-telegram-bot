package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"github.com/maddevsio/simple-config"
)

var (
	config = simple_config.NewSimpleConfig("config", "yml")
)


func main() {
	bot, err := tgbotapi.NewBotAPI(config.GetString("bot_token"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Быстрый заказ такси"),
		),
		//tgbotapi.NewKeyboardButtonRow(
		//	tgbotapi.NewKeyboardButton("Заказ такси"),
		//	tgbotapi.NewKeyboardButton("Машины рядом"),
		//),
		//tgbotapi.NewKeyboardButtonRow(
		//	tgbotapi.NewKeyboardButton("Тарифы"),
		//	tgbotapi.NewKeyboardButton("Помощь"),
		//),
	)

	keyboard.OneTimeKeyboard = true

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.Text == "Быстрый заказ такси" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Укажите ваш телефон")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Что-что?")
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}
	}
}