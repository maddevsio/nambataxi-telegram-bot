package chat

import "github.com/maddevsio/nambataxi-telegram-bot/api"

type Session struct {
	Phone string
	Address string
	FareId int
	State string
	Order api.Order
}

const (
	STATE_NEED_PHONE    = "need phone"
	STATE_NEED_ADDRESS  = "need address"
	STATE_NEED_FARE     = "need fare"
	STATE_ORDER_CREATED = "order created"
)
