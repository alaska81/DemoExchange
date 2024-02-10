package entities

type Wallet struct {
	Exchange   Exchange   `json:"exchange" db:"exchange"`
	AccountUID AccountUID `json:"account_uid" db:"account_uid"`
	Balance    Balance    `json:"balance"`
	UpdateTS   int64      `json:"update_ts" db:"update_ts"`
}

type Balance struct {
	Coin  Coin    `json:"coin" db:"coin"`
	Total float64 `json:"total" db:"total"`
	Hold  float64 `json:"hold" db:"hold"`

	WalletBalance float64 `json:"wallet_balance" db:"wallet_balance"`
	MarginBalance float64 `json:"margin_balance" db:"margin_balance"`
	InitialMargin float64 `json:"initial_margin" db:"initial_margin"`
	MaintMargin   float64 `json:"maint_margin" db:"maint_margin"`
}

type Coin string

type Balances map[Coin]Balance
