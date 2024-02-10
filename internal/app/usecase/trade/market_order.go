package trade

import (
	"context"
	"fmt"

	"DemoExchange/internal/app/entities"
)

type MarketOrder struct {
	origin   *entities.Order
	Exchange entities.Exchange
	Symbol   entities.Symbol
	Side     entities.OrderSide
	Amount   float64
	Price    float64
}

func NewMarketOrder(order *entities.Order) *MarketOrder {
	return &MarketOrder{
		origin:   order,
		Exchange: order.Exchange,
		Symbol:   order.Symbol,
		Side:     order.Side,
		Amount:   order.Amount,
		Price:    order.Price,
	}
}

func (o *MarketOrder) Validate() error {
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

	return nil
}

func (o *MarketOrder) SetPrice(price float64) {
	o.origin.Price = price
	o.Price = o.origin.Price
}

func (o *MarketOrder) Process(ctx context.Context) <-chan *entities.Order {
	ch := make(chan *entities.Order)

	go func() {
		defer func() {
			close(ch)
			fmt.Println("MarketOrder.Process: close", o.origin)
		}()

		fmt.Printf("MarketOrder.Process: %+v\n", o.origin)
		o.origin.Status = entities.OrderStatusPending
		ch <- o.origin

		for {
			select {
			case <-ctx.Done():
				return
			default:
				o.origin.Status = entities.OrderStatusSuccess
				ch <- o.origin
				return
			}
		}
	}()

	return ch
}
