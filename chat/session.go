package chat

import "github.com/maddevsio/nambataxi-telegram-bot/api"

type Session struct {
	Phone string
	Address string
	FareId int
	State string
	Order api.Order
}
