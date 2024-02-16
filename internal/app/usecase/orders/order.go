package orders

import (
	"context"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

type Order struct {
	order *entities.Order
}

func NewOrder(ctx context.Context, uc Usecase, o *entities.Order) (*Order, error) {
	if o.Exchange == entities.ExchangeFutures {
		account, err := uc.GetAccountByUID(ctx, o.AccountUID)
		if err != nil {
			return nil, err
		}
		o.PositionMode = account.PositionMode
	}

	return &Order{o}, nil
}

func (o *Order) GetOrder() *entities.Order {
	return o.order
}

func (o *Order) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.Exchange {
	case entities.ExchangeSpot:
		return NewOrderSpot(o.order).HoldBalance(ctx, uc, log)
	case entities.ExchangeFutures:
		return NewOrderFutures(o.order).HoldBalance(ctx, uc, log)
	default:
		return apperror.ErrExchangeIsNotValid
	}
}

func (o *Order) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.Exchange {
	case entities.ExchangeSpot:
		return NewOrderSpot(o.order).UnholdBalance(ctx, uc, log)
	case entities.ExchangeFutures:
		return NewOrderFutures(o.order).UnholdBalance(ctx, uc, log)
	default:
		return apperror.ErrExchangeIsNotValid
	}
}

func (o *Order) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.Exchange {
	case entities.ExchangeSpot:
		return NewOrderSpot(o.order).AppendBalance(ctx, uc, log)
	case entities.ExchangeFutures:
		return NewOrderFutures(o.order).AppendBalance(ctx, uc, log)
	default:
		return apperror.ErrExchangeIsNotValid
	}
}

func (o *Order) Validate() error {
	if o.order.Exchange != entities.ExchangeSpot && o.order.Exchange != entities.ExchangeFutures {
		return apperror.ErrExchangeIsNotValid
	}

	if o.order.Symbol == "" {
		return apperror.ErrSymbolIsNotValid
	}

	if o.order.Side != entities.OrderSideBuy && o.order.Side != entities.OrderSideSell {
		return apperror.ErrOrderSideIsNotValid
	}

	if o.order.Amount <= 0 {
		return apperror.ErrAmountIsNotValid
	}

	var err error

	switch o.order.Type {
	case entities.OrderTypeMarket:
		err = NewMarketOrder(o.order).Validate()

	case entities.OrderTypeLimit:
		err = NewLimitOrder(o.order).Validate()

	default:
		return apperror.ErrOrderTypeNotValid
	}

	if err != nil {
		return err
	}

	if o.order.Exchange == entities.ExchangeFutures {
		return NewOrderFutures(o.order).Validate()
	}

	return nil
}

func (o *Order) Process(ctx context.Context) (<-chan entities.OrderStatus, error) {
	switch o.order.Type {
	case entities.OrderTypeMarket:
		return NewMarketOrder(o.order).Process(ctx), nil

	case entities.OrderTypeLimit:
		return NewLimitOrder(o.order).Process(ctx), nil

	default:
		return nil, apperror.ErrOrderTypeNotValid
	}
}
