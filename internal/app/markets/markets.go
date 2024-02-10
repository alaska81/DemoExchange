package markets

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrMarketsIsEmpty = errors.New("Markets is empty")
	ErrMarketNotFound = errors.New("Market not found")

// ErrTickerNotValid = errors.New("Ticker not valid")
)

type Receiver struct{}

func New() *Receiver {
	return &Receiver{}
}

func (r *Receiver) GetMarkets(exchange string) (markets Markets, err error) {
	if v, ok := marketsMap.Load(exchange); ok {
		markets = v.(Markets)
		return
	}

	err = ErrMarketsIsEmpty
	return
}

func (r *Receiver) GetMarket(exchange, market string) (Market, error) {
	var result Market

	markets, err := r.GetMarkets(exchange)
	if err != nil {
		return result, err
	}

	result, ok := markets[market]
	if !ok {
		return result, ErrMarketNotFound
	}

	return result, nil
}

func (r *Receiver) GetMarketWithContext(ctx context.Context, exchange, market string) (Market, error) {
	var result Market

	for {
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("Market is error (%s, %s): %s", exchange, market, ctx.Err())
		default:
			var err error
			result, err = r.GetMarket(exchange, market)
			if err != nil {
				time.Sleep(timeout)
				continue
			}
			return result, nil
		}
	}
}
