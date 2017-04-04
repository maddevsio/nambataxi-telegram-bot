package chat

import (
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/jinzhu/gorm"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"time"
	"log"
)

func StartStatusReactionGoroutine(nambaTaxiAPI api.NambaTaxiApi, update tgbotapi.Update, bot *tgbotapi.BotAPI, db *gorm.DB, session storage.Session) {
	go func(session storage.Session) {
		var status = "Новый заказ"
		for {
			time.Sleep(5 * time.Second)
			log.Printf("Session order id: %v", session.OrderId)
			currentOrder, err := nambaTaxiAPI.GetOrder(session.OrderId)
			if err != nil {
				log.Printf("Error getting order status %v", err)
				return
			} else {
				log.Printf("Order status %v", currentOrder.Status)
			}
			if status != currentOrder.Status {
				OrderStatusReaction(currentOrder, update, bot, db, session)
			}
			status = currentOrder.Status
			if currentOrder.Status == "Отклонен" || currentOrder.Status == "Выполнен" {
				return
			}
		}
	}(session)
}

func OrderStatusReaction(order api.Order, update tgbotapi.Update, bot *tgbotapi.BotAPI, db *gorm.DB, session storage.Session) {
	basicKeyboard := GetBasicKeyboard()
	orderKeyboard := GetOrderKeyboard()

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
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, BOT_ORDER_DONE)
		msg.ReplyMarkup = basicKeyboard
		bot.Send(msg)
		return
	}

	if order.Status == "Отклонен" {
		db.Delete(&session)
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%v", order.Status))
		msg.ReplyMarkup = basicKeyboard
		bot.Send(msg)
		return
	}

	msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%v", order.Status))
	msg.ReplyMarkup = orderKeyboard
	bot.Send(msg)
	return
}
