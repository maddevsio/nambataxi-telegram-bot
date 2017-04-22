package holder

import (
	"gopkg.in/telegram-bot-api.v4"
	"github.com/jinzhu/gorm"
	"github.com/maddevsio/nambataxi-telegram-bot/api"
	"github.com/maddevsio/simple-config"
)

type Service struct {
	Update       tgbotapi.Update
	Bot          *tgbotapi.BotAPI
	DB           *gorm.DB
	NambaTaxiAPI api.NambaTaxiAPI
	Config       simple_config.SimpleConfig
}