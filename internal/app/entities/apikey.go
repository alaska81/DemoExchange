package entities

import (
	"time"

	"DemoExchange/internal/app/pkg/hash"
)

type Key struct {
	Token      Token      `json:"token" db:"token"`
	AccountUID AccountUID `json:"account_uid" db:"account_uid"`
	Disabled   bool       `json:"disabled" db:"disabled"`
	CreateTS   int64      `json:"create_ts" db:"create_ts"`
	UpdateTS   int64      `json:"update_ts" db:"update_ts"`
}

type Token string

func NewToken(accountUID AccountUID) *Key {
	ts := TS()

	return &Key{
		Token:      Token(hash.GenSHA1(string(accountUID) + time.Now().String())),
		AccountUID: accountUID,
		CreateTS:   ts,
		UpdateTS:   ts,
	}
}
