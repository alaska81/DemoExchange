package entities

import (
	"time"

	"github.com/google/uuid"
)

const (
	DefaultPositionMode = PositionModeOneway
)

type Account struct {
	AccountUID   AccountUID   `json:"account_uid" db:"account_uid"`
	Service      string       `json:"service" db:"service"`
	UserID       string       `json:"user_id" db:"user_id"`
	PositionMode PositionMode `json:"position_mode" db:"position_mode"`
	CreateTS     int64        `json:"create_ts" db:"create_ts"`
	UpdateTS     int64        `json:"update_ts" db:"update_ts"`
	IsNew        bool         `json:"-" db:"-"`
}

type AccountUID string

func NewAccount(service, userID string) *Account {
	ts := TS()

	return &Account{
		AccountUID:   AccountUID(uuid.New().String()),
		Service:      service,
		UserID:       userID,
		PositionMode: DefaultPositionMode,
		CreateTS:     ts,
		UpdateTS:     ts,
	}
}

func TS() int64 {
	return time.Now().UTC().UnixMilli()
}
