package orders

import (
	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
	"context"
)

type OrderSpot struct {
	order *entities.Order
}

func NewOrderSpot(o *entities.Order) *OrderSpot {
	return &OrderSpot{o}
}

func (o *OrderSpot) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.Side {
	case entities.OrderSideBuy:
		return NewOrderSpotBuy(o.order).HoldBalance(ctx, uc, log)
	case entities.OrderSideSell:
		return NewOrderSpotSell(o.order).HoldBalance(ctx, uc, log)
	default:
		return apperror.ErrOrderSideIsNotValid
	}
}

func (o *OrderSpot) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.Side {
	case entities.OrderSideBuy:
		return NewOrderSpotBuy(o.order).UnholdBalance(ctx, uc, log)
	case entities.OrderSideSell:
		return NewOrderSpotSell(o.order).UnholdBalance(ctx, uc, log)
	default:
		return apperror.ErrOrderSideIsNotValid
	}
}

func (o *OrderSpot) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.Side {
	case entities.OrderSideBuy:
		return NewOrderSpotBuy(o.order).AppendBalance(ctx, uc, log)
	case entities.OrderSideSell:
		return NewOrderSpotSell(o.order).AppendBalance(ctx, uc, log)
	default:
		return apperror.ErrOrderSideIsNotValid
	}
}

func (o *OrderSpot) Validate(ctx context.Context, markets Markets) error {
	market, err := markets.GetMarketWithContext(context.Background(), o.order.Exchange.Name(), o.order.Symbol.String())
	if err != nil {
		return err
	}

	limits := market.Limits.Amount

	if limits.Min > 0 && o.order.Amount < limits.Min {
		return apperror.ErrAmountIsOutOfRange
	}

	return nil
}
