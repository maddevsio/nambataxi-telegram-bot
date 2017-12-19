package holder

import (
	"github.com/jinzhu/gorm"
	"github.com/maddevsio/nambataxi-api-go-client/api"
	"github.com/maddevsio/simple-config"
	"gopkg.in/telegram-bot-api.v4"
)

type Service struct {
	Update       tgbotapi.Update
	Bot          *tgbotapi.BotAPI
	DB           *gorm.DB
	NambaTaxiAPI api.NambaTaxiAPI
	Config       simple_config.SimpleConfig
}
