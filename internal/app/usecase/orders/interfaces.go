package orders

import (
	"context"

	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/markets"
	"DemoExchange/internal/app/tickers"
)

type Account interface {
	GetAccountByUID(ctx context.Context, accountUID entities.AccountUID) (*entities.Account, error)
}

type Balance interface {
	GetBalanceCoin(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin) (total float64, hold float64, err error)
	SetHoldBalance(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, balance entities.Balance) error
	SubtractBalance(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, balance entities.Balance) error
	AppendBalance(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, balance entities.Balance) error
}

type Position interface {
	GetPositionBySide(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol entities.Symbol, side entities.PositionSide) (*entities.Position, error)
	SavePosition(ctx context.Context, position *entities.Position) error
}

type Usecase interface {
	Account
	Balance
	Position
}

type Tickers interface {
	GetTickerWithContext(ctx context.Context, exchange, market string) (tickers.Ticker, error)
}

type Markets interface {
	GetMarketWithContext(ctx context.Context, exchange, market string) (markets.Market, error)
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type Storage interface {
	Set(order *entities.Order)
	Get(orderUID string) (*entities.Order, bool)
	Delete(orderUID string)
	List() []*entities.Order
}
