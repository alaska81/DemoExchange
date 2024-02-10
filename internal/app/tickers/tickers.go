package tickers

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTickersIsEmpty = errors.New("Tickers is empty")
	ErrTickerNotFound = errors.New("Ticker not found")
	ErrTickerNotValid = errors.New("Ticker not valid")
)

type Receiver struct{}

func New() *Receiver {
	return &Receiver{}
}

func (r *Receiver) GetTickers(exchange string) (tickers Tickers, err error) {
	if v, ok := tickersMap.Load(exchange); ok {
		tickers = v.(Tickers)
		return
	}

	err = ErrTickersIsEmpty
	return
}

func (r *Receiver) GetTicker(exchange, market string) (Ticker, error) {
	var ticker Ticker

	tickers, err := r.GetTickers(exchange)
	if err != nil {
		return ticker, err
	}

	ticker, ok := tickers[market]
	if !ok {
		return ticker, ErrTickerNotFound
	}

	// if ticker.Ask == 0 || ticker.Bid == 0 {
	// 	return ticker, ErrTickerNotValid
	// 	// if _, err = markets.GetMarket(exchange, market); err != nil {
	// 	// 	return
	// 	// }
	// }

	if ticker.Last == 0 {
		return ticker, ErrTickerNotValid
	}

	return ticker, nil
}

func (r *Receiver) GetTickerWithContext(ctx context.Context, exchange, market string) (Ticker, error) {
	var ticker Ticker

	for {
		select {
		case <-ctx.Done():
			return ticker, fmt.Errorf("Ticker is error (%s, %s): %s", exchange, market, ctx.Err())
		default:
			var err error
			ticker, err = r.GetTicker(exchange, market)
			if err != nil {
				time.Sleep(timeout)
				continue
			}
			return ticker, nil
		}
	}
}
