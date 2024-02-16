package orders

import (
	"DemoExchange/internal/app/entities"
	"context"
)

type OrderFuturesHedge struct {
	order *entities.Order
}

func NewOrderFuturesHedge(o *entities.Order) *OrderFuturesHedge {
	return &OrderFuturesHedge{o}
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
