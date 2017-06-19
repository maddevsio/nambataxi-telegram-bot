package chat

import (
	"gopkg.in/telegram-bot-api.v4"
	"fmt"
	"log"
	"errors"
	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
)

var NambaTaxiApi api.NambaTaxiAPI

func GetBasicKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BOT_MESSAGE_ORDER_FAST_START),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BOT_MESSAGE_FARES),
		),
	)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func GetPhoneKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact(BOT_MESSAGE_SEND_MY_PHONE),
		),
	)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func GetOrderKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BOT_MESSAGE_MY_ORDER_STATUS),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BOT_MESSAGE_NEAREST_CARS),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BOT_MESSAGE_CANCEL),
		),
	)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func GetFaresKeyboard() tgbotapi.ReplyKeyboardMarkup {
	fares, err := NambaTaxiApi.GetFares()
	if err != nil {
		log.Printf("error getting fares: %v", err)
		return tgbotapi.NewReplyKeyboard()
	}

	var rows [][]tgbotapi.KeyboardButton
	for _, fare := range fares.Fare {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fare.Name)))
	}

	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func GetAddressKeyboard(addresses []storage.Address) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, address := range addresses {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(address.Text)))
	}
	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func GetPhonesKeyboard(phones []storage.Phone) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, phone := range phones {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(phone.Text)))
	}
	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.OneTimeKeyboard = true
	return keyboard
}

func GetFareIdByName(fareName string) (int, error) {
	fares, err := NambaTaxiApi.GetFares()
	if err != nil {
		log.Printf("error getting fares: %v", err)
		return 0, err
	}
	for _, fare := range fares.Fare {
		if fare.Name == fareName {
			return fare.ID, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("Cannot find fare with name %v", fareName))
}
