package main

import (
	"log"

	"strconv"

	"github.com/maddevsio/nambataxi-api-go-client/api"
	"github.com/maddevsio/nambataxi-telegram-bot/chat"
	"github.com/maddevsio/nambataxi-telegram-bot/holder"
	"github.com/maddevsio/nambataxi-telegram-bot/storage"
	"github.com/maddevsio/simple-config"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	service holder.Service
	err     error
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
