package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) holdBalanceFutures(ctx context.Context, order *entities.Order) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if order.PositionSide == entities.PositionSideBoth {
			return uc.holdBalanceFuturesOneway(ctx, tx, order)
		} else {
			return uc.holdBalanceFuturesHedge(ctx, tx, order)
		}
	})
}

func (uc *Usecase) unholdBalanceFutures(ctx context.Context, order *entities.Order) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if order.PositionSide == entities.PositionSideBoth {
			return uc.unholdBalanceFuturesOneway(ctx, tx, order)
		} else {
			return uc.unholdBalanceFuturesHedge(ctx, tx, order)
		}
	})
}

func (uc *Usecase) appendBalanceFutures(ctx context.Context, order *entities.Order) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if order.PositionSide == entities.PositionSideBoth {
			return uc.appendBalanceFuturesOneway(ctx, tx, order)
		} else {
			return uc.appendBalanceFuturesHedge(ctx, tx, order)
		}
	})
}
