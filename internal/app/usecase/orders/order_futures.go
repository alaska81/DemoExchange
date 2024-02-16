package orders

import (
	"context"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

type OrderFutures struct {
	order *entities.Order
}

func NewOrderFutures(o *entities.Order) *OrderFutures {
	return &OrderFutures{o}
}

func (o *OrderFutures) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.PositionMode {
	case entities.PositionModeOneway:
		return NewOrderFuturesOneway(o.order).HoldBalance(ctx, uc, log)
	case entities.PositionModeHedge:
		return NewOrderFuturesHedge(o.order).HoldBalance(ctx, uc, log)
	default:
		return apperror.ErrOrderPositionModeIsNotValid
	}
}

func (o *OrderFutures) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.PositionMode {
	case entities.PositionModeOneway:
		return NewOrderFuturesOneway(o.order).UnholdBalance(ctx, uc, log)
	case entities.PositionModeHedge:
		return NewOrderFuturesHedge(o.order).UnholdBalance(ctx, uc, log)
	default:
		return apperror.ErrOrderPositionModeIsNotValid
	}
}

func (o *OrderFutures) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	switch o.order.PositionMode {
	case entities.PositionModeOneway:
		return NewOrderFuturesOneway(o.order).AppendBalance(ctx, uc, log)
	case entities.PositionModeHedge:
		return NewOrderFuturesHedge(o.order).AppendBalance(ctx, uc, log)
	default:
		return apperror.ErrOrderPositionModeIsNotValid
	}
}

func (o *OrderFutures) Validate() error {
	if o.order.PositionMode == entities.PositionModeOneway {
		o.order.PositionSide = entities.PositionSideBoth
	} else {
		if o.order.PositionSide != entities.PositionSideLong && o.order.PositionSide != entities.PositionSideShort {
			return apperror.ErrInvalidPositionSide
		}
	}

	return nil
}
