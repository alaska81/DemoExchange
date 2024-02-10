package trade

import (
	"context"

	"DemoExchange/internal/app/tickers"
)

type Tickers interface {
	GetTickerWithContext(ctx context.Context, exchange, market string) (tickers.Ticker, error)
}
