package entities

import (
	"strings"

	"github.com/google/uuid"
)

type Order struct {
	AccountUID   AccountUID   `json:"account_uid" db:"account_uid"`
	OrderUID     string       `json:"order_uid" db:"order_uid"`
	Exchange     Exchange     `json:"exchange" db:"exchange"`
	Symbol       Symbol       `json:"symbol" db:"symbol"`
	Type         OrderType    `json:"type" db:"type"`
	Side         OrderSide    `json:"side" db:"side"`
	Amount       float64      `json:"amount" db:"amount"`
	Price        float64      `json:"price" db:"price"`
	Status       OrderStatus  `json:"status" db:"status"`
	Error        string       `json:"error,omitempty" db:"error"`
	CreateTS     int64        `json:"create_ts" db:"create_ts"`
	UpdateTS     int64        `json:"update_ts" db:"update_ts"`
	PositionSide PositionSide `json:"position_side" db:"position_side"`
	PositionMode PositionMode `json:"position_mode" db:"position_mode"`
}

type Orders []Order

type Symbol string

func (s Symbol) String() string {
	return string(s)
}

type Coins struct {
	CoinQuote Coin
	CoinBase  Coin
}

func (s Symbol) GetCoins() Coins {
	coins := strings.Split(string(s), "/")
	if len(coins) != 2 {
		return Coins{}
	}
	return Coins{Coin(coins[0]), Coin(coins[1])}
}

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "new"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusSuccess   OrderStatus = "done"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusFailed    OrderStatus = "failed"
)

func (s OrderStatus) String() string {
	return string(s)
}

func StatusArrayToString(statuses []OrderStatus) string {
	var statusArray []string
	for _, status := range statuses {
		statusArray = append(statusArray, status.String())
	}
	return strings.Join(statusArray, ",")
}

type OrderType string

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

func NewOrder(accountUID AccountUID) *Order {
	ts := TS()

	return &Order{
		AccountUID: accountUID,
		OrderUID:   uuid.New().String(),
		Status:     OrderStatusNew,
		CreateTS:   ts,
		UpdateTS:   ts,
	}
}
