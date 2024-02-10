package trade

import (
	"context"
	"fmt"
	"time"

	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/tickers"
)

type LimitOrder struct {
	origin   *entities.Order
	Exchange entities.Exchange
	Symbol   entities.Symbol
	Side     entities.OrderSide
	Amount   float64
	Price    float64
	tickers  Tickers
}

func NewLimitOrder(order *entities.Order) *LimitOrder {
	return &LimitOrder{
		origin:   order,
		Exchange: order.Exchange,
		Symbol:   order.Symbol,
		Side:     order.Side,
		Amount:   order.Amount,
		Price:    order.Price,
		tickers:  tickers.New(),
	}
}

func (o *LimitOrder) Validate() error {
	if o.Exchange != entities.ExchangeSpot && o.Exchange != entities.ExchangeFutures {
		return ErrExchangeIsNotValid
	}

	if o.Symbol == "" {
		return ErrSymbolIsNotValid
	}

	if o.Side != entities.OrderSideBuy && o.Side != entities.OrderSideSell {
		return ErrOrderSideIsNotValid
	}

	if o.Amount <= 0 {
		return ErrAmountIsNotValid
	}

	if o.Price <= 0 {
		return ErrPriceIsNotValid
	}

	return nil
}

func (o *LimitOrder) Process(ctx context.Context) <-chan *entities.Order {
	ch := make(chan *entities.Order)

	go func() {
		defer func() {
			close(ch)
			fmt.Println("LimitOrder.Process: close", o.origin)
		}()

		fmt.Println("LimitOrder.Process", o.origin)
		o.origin.Status = entities.OrderStatusPending
		ch <- o.origin

		var (
			ticker tickers.Ticker
			err    error
		)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if o.origin.Status == entities.OrderStatusSuccess || o.origin.Status == entities.OrderStatusCancelled {
					ch <- o.origin
					return
				}

				ticker, err = o.tickers.GetTickerWithContext(ctx, o.Exchange.Name(), o.Symbol.String())
				if err != nil {
					time.Sleep(processTimeout)
					continue
				}

				if o.origin.Status == entities.OrderStatusSuccess || o.origin.Status == entities.OrderStatusCancelled {
					ch <- o.origin
					return
				}

				if (o.Side == "buy" && ticker.Last <= o.Price) || (o.Side == "sell" && ticker.Last >= o.Price) {
					o.origin.Status = entities.OrderStatusSuccess
				}

				// fmt.Println("Process.order: ", o.origin, "ticker: ", ticker.Last)

				if o.origin.Status == entities.OrderStatusSuccess || o.origin.Status == entities.OrderStatusCancelled {
					ch <- o.origin
					return
				}

				time.Sleep(processTimeout)
			}
		}
	}()

	return ch
}
