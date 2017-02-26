package chat

type Session struct {
	Phone string
	Address string
	FareId int
	State string
	OrderId int
}

const (
	STATE_NEED_PHONE    = "need phone"
	STATE_NEED_ADDRESS  = "need address"
	STATE_NEED_FARE     = "need fare"
	STATE_ORDER_CREATED = "order created"
)

func GetAllSessions() map[int64]*Session{
	return make(map[int64]*Session)
	// TODO need to get this data from SQLite3 via GORM
}