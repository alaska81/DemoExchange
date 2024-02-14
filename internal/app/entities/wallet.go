package entities

type Wallet struct {
	Exchange   Exchange   `json:"exchange" db:"exchange"`
	AccountUID AccountUID `json:"account_uid" db:"account_uid"`
	Balance    Balance    `json:"balance"`
	UpdateTS   int64      `json:"update_ts" db:"update_ts"`
}

type Coin string

type Balance struct {
	Coin  Coin    `json:"coin" db:"coin"`
	Total float64 `json:"total" db:"total"`
	Hold  float64 `json:"hold" db:"hold"`

	AvailableBalance float64 `json:"available_balance" db:"-"`
	WalletBalance    float64 `json:"wallet_balance" db:"-"`
	MarginBalance    float64 `json:"margin_balance" db:"-"`
	InitialMargin    float64 `json:"initial_margin" db:"-"`
	MaintMargin      float64 `json:"maint_margin" db:"-"`
	UnrealisedPnl    float64 `json:"unrealised_pnl" db:"-"`
}

type Balances map[Coin]Balance
