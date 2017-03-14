package chat

import (
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/jinzhu/gorm"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
)

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
