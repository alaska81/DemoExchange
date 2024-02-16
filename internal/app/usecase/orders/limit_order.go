package orders

import (
	"context"
	"fmt"
	"time"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/tickers"
)

const processTimeout = 5 * time.Second

type LimitOrder struct {
	order   *entities.Order
	tickers Tickers
}

func NewLimitOrder(o *entities.Order) *LimitOrder {
	return &LimitOrder{
		order:   o,
		tickers: tickers.New(),
	}
}

func (o *LimitOrder) Validate() error {
	if o.order.Price <= 0 {
		return apperror.ErrPriceIsNotValid
	}

	return nil
}

func (o *LimitOrder) Process(ctx context.Context) <-chan entities.OrderStatus {
	ch := make(chan entities.OrderStatus)

	go func() {
		defer func() {
			close(ch)
			fmt.Println("LimitOrder.Process: close", o.order)
		}()

		fmt.Println("LimitOrder.Process", o.order)
		o.order.Status = entities.OrderStatusPending
		ch <- entities.OrderStatusPending

		var (
			ticker tickers.Ticker
			err    error
		)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				status := o.order.Status
				if status == entities.OrderStatusSuccess || status == entities.OrderStatusCancelled {
					ch <- status
					return
				}

				ticker, err = o.tickers.GetTickerWithContext(ctx, o.order.Exchange.Name(), o.order.Symbol.String())
				if err != nil {
					time.Sleep(processTimeout)
					continue
				}

				status = o.order.Status
				if status == entities.OrderStatusSuccess || status == entities.OrderStatusCancelled {
					ch <- status
					return
				}

				if (o.order.Side == "buy" && ticker.Last <= o.order.Price) || (o.order.Side == "sell" && ticker.Last >= o.order.Price) {
					ch <- entities.OrderStatusSuccess
					return
				}

				time.Sleep(processTimeout)
			}
		}
	}()

	return ch
}
