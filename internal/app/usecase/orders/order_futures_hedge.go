package orders

import (
	"context"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

type OrderFuturesHedge struct {
	order *entities.Order
}

func NewOrderFuturesHedge(o *entities.Order) *OrderFuturesHedge {
	return &OrderFuturesHedge{o}
}

func (o *OrderFuturesHedge) Validate() error {
	if o.order.PositionSide != entities.PositionSideLong && o.order.PositionSide != entities.PositionSideShort {
		return apperror.ErrInvalidPositionSide
	}

	if (o.order.PositionSide == entities.PositionSideLong && o.order.Side == entities.OrderSideBuy) || (o.order.PositionSide == entities.PositionSideShort && o.order.Side == entities.OrderSideSell) {
		return NewOrderFuturesHedgeOpen(o.order).Validate()
	} else {
		return NewOrderFuturesHedgeClose(o.order).Validate()
	}
}

func (o *OrderFuturesHedge) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	if (o.order.PositionSide == entities.PositionSideLong && o.order.Side == entities.OrderSideBuy) || (o.order.PositionSide == entities.PositionSideShort && o.order.Side == entities.OrderSideSell) {
		return NewOrderFuturesHedgeOpen(o.order).HoldBalance(ctx, uc, log)
	} else {
		return NewOrderFuturesHedgeClose(o.order).HoldBalance(ctx, uc, log)
	}
}

func (o *OrderFuturesHedge) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	if (o.order.PositionSide == entities.PositionSideLong && o.order.Side == entities.OrderSideBuy) || (o.order.PositionSide == entities.PositionSideShort && o.order.Side == entities.OrderSideSell) {
		return NewOrderFuturesHedgeOpen(o.order).UnholdBalance(ctx, uc, log)
	} else {
		return NewOrderFuturesHedgeClose(o.order).UnholdBalance(ctx, uc, log)
	}
}

func (o *OrderFuturesHedge) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	if (o.order.PositionSide == entities.PositionSideLong && o.order.Side == entities.OrderSideBuy) || (o.order.PositionSide == entities.PositionSideShort && o.order.Side == entities.OrderSideSell) {
		return NewOrderFuturesHedgeOpen(o.order).AppendBalance(ctx, uc, log)
	} else {
		return NewOrderFuturesHedgeClose(o.order).AppendBalance(ctx, uc, log)
	}
}
