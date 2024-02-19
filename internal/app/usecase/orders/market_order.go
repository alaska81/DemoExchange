package orders

import (
	"context"
	"fmt"
	"time"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/tickers"
)

type MarketOrder struct {
	order   *entities.Order
	tickers Tickers
}

func NewMarketOrder(o *entities.Order) *MarketOrder {
	return &MarketOrder{
		order:   o,
		tickers: tickers.New(),
	}
}

func (o *MarketOrder) Validate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker, err := o.tickers.GetTickerWithContext(ctx, o.order.Exchange.Name(), o.order.Symbol.String())
	if err != nil {
		return apperror.ErrMarketPriceIsWrong
	}

	o.order.Price = ticker.Last

	return nil
}

func (o *MarketOrder) Process(ctx context.Context) <-chan entities.OrderStatus {
	ch := make(chan entities.OrderStatus)

	go func() {
		defer func() {
			close(ch)
			fmt.Printf("MarketOrder.Process:close %+v\n", o.order)
		}()

		fmt.Printf("MarketOrder.Process: %+v\n", o.order)
		ch <- entities.OrderStatusPending

		for {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- entities.OrderStatusSuccess
				return
			}
		}
	}()

	return ch
}
