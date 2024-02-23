package entities

import "github.com/google/uuid"

type Transaction struct {
	AccountUID      AccountUID      `json:"account_uid" db:"account_uid"`
	TransactionID   string          `json:"transaction_id" db:"transaction_id"`
	Exchange        Exchange        `json:"exchange" db:"exchange"`
	Symbol          Symbol          `json:"symbol" db:"symbol"`
	TransactionType TransactionType `json:"transaction_type" db:"transaction_type"`
	Amount          float64         `json:"amount" db:"amount"`
	CreateTS        int64           `json:"create_ts" db:"create_ts"`
	// TradeUID      string     `json:"trade_uid" db:"trade_uid"`
}

func NewTransaction(accountUID AccountUID, exchange Exchange, symbol Symbol, transactionType TransactionType, amount float64) *Transaction {
	return &Transaction{
		AccountUID:      accountUID,
		TransactionID:   uuid.New().String(),
		Exchange:        exchange,
		Symbol:          symbol,
		TransactionType: transactionType,
		Amount:          amount,
		CreateTS:        TS(),
	}
}

type TransactionType string

const (
	TransactionTypeLiquidation TransactionType = "liquidation"
)

type TransactionFilter struct {
	TransactionType string `db:"transaction_type"`
	From            int64  `db:"from"`
	To              int64  `db:"to"`
	Limit           int64  `db:"limit"`
}
