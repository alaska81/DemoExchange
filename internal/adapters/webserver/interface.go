package webserver

import (
	"context"

	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/markets"
	"DemoExchange/internal/app/orderbook"
	"DemoExchange/internal/app/tickers"
)

type Markets interface {
	GetMarkets(exchange string) (markets markets.Markets, err error)
}

type Tickers interface {
	GetTickers(exchange string) (tickers tickers.Tickers, err error)
}

type Orderbook interface {
	GetOrderbook(ctx context.Context, exchange, symbol, limit string) (*orderbook.Orderbook, error)
}

type Usecase interface {
	SetAccountPositionMode(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, positionMode entities.PositionMode) error

	CreateToken(ctx context.Context, service, userID string) (entities.Token, error)
	DisableToken(ctx context.Context, token entities.Token) error
	GetAccountUID(ctx context.Context, token entities.Token) (entities.AccountUID, error)

	GetBalances(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) (entities.Balances, error)
	Deposit(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) (float64, error)
	Withdraw(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) error

	NewOrder(ctx context.Context, order *entities.Order) error
	GetOrder(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, orderUID string) (*entities.Order, error)
	CancelOrder(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, orderUID string) (*entities.Order, error)
	OrdersList(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, statuses []entities.OrderStatus, limit int) ([]*entities.Order, error)

	PositionsList(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) ([]*entities.Position, error)
	SetPositionMarginType(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, marginType entities.MarginType) error
	SetPositionLeverage(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, leverage entities.PositionLeverage) error

	TransactionsList(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, filter entities.TransactionFilter) ([]*entities.Transaction, error)
}

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Trace(args ...interface{})
	Tracef(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}
