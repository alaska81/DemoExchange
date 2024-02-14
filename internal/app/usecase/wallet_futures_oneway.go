package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) holdBalanceFuturesOneway(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	position, err := uc.getPositionBySide(ctx, tx, order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesOneway:getPositionBySide [%+v] error: %v", order, err))
		return err
	}

	balancePosition := 0.0
	if order.Side == entities.OrderSideBuy && position.Amount < 0 {
		balancePosition = -position.Amount - position.HoldAmount
	} else if order.Side == entities.OrderSideSell && position.Amount > 0 {
		balancePosition = position.Amount - position.HoldAmount
	}

	holdPosition := order.Amount
	if holdPosition > balancePosition {
		holdPosition = balancePosition
	}

	if holdPosition > 0 {
		position.HoldAmount += holdPosition
		position.UpdateTS = order.UpdateTS

		err = uc.position.UpdatePosition(ctx, tx, position)
		if err != nil {
			uc.log.Error(fmt.Sprintf("holdBalanceFuturesOneway:UpdatePosition [%+v] error: %v", position, err))
			return err
		}
	}

	hold := (order.Amount - holdPosition) * order.Price
	if hold > 0 {
		coins := order.Symbol.GetCoins()
		coin := coins.CoinBase

		balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
		if err != nil {
			uc.log.Error(fmt.Sprintf("holdBalanceFuturesOneway:getBalanceCoin [%+v] error: %v", order, err))
			return err
		}

		if hold > balanceTotal-balanceHold {
			uc.log.Error(fmt.Sprintf("holdBalanceFuturesOneway:ErrInsufficientFunds [AccountUID: %s, exchange: %s, coin: %s, balance_total: %v, balance_hold: %v, hold: %v]", order.AccountUID, order.Exchange, coin, balanceTotal, balanceHold, hold))
			return apperror.ErrInsufficientFunds
		}

		wallet := entities.Wallet{
			Exchange:   order.Exchange,
			AccountUID: order.AccountUID,
			Balance: entities.Balance{
				Coin: coin,
				Hold: hold,
			},
			UpdateTS: entities.TS(),
		}

		err = uc.wallet.SetHoldCoin(ctx, tx, wallet)
		if err != nil {
			uc.log.Error(fmt.Sprintf("holdBalanceFuturesOneway:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, hold: %v] error: %v", order.AccountUID, order.Exchange, coin, hold, err))
			return err
		}
	}

	return nil
}

func (uc *Usecase) unholdBalanceFuturesOneway(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins := order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceFuturesOneway:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	cost := order.Amount * order.Price
	unhold := balanceHold - cost
	if unhold > balanceTotal {
		unhold = balanceTotal
	}

	if unhold < 0 {
		unhold = 0
	}

	wallet := entities.Wallet{
		Exchange:   order.Exchange,
		AccountUID: order.AccountUID,
		Balance: entities.Balance{
			Coin: coin,
			Hold: unhold,
		},
		UpdateTS: entities.TS(),
	}

	err = uc.wallet.SetHoldCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceFuturesHedge:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, unhold: %v] error: %v", order.AccountUID, order.Exchange, coin, unhold, err))
		return err
	}

	position, err := uc.getPositionBySide(ctx, tx, order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceFuturesOneway:getPositionBySide [%+v] error: %v", order, err))
		return err
	}

	position.HoldAmount -= (order.Amount - (balanceHold+unhold)/order.Price)

	if position.HoldAmount > position.Amount {
		position.HoldAmount = position.Amount
	}

	if position.HoldAmount < 0 {
		position.HoldAmount = 0
	}

	err = uc.position.UpdatePosition(ctx, tx, position)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceFuturesOneway:UpdatePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}

func (uc *Usecase) appendBalanceFuturesOneway(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins := order.Symbol.GetCoins()
	coin := coins.CoinBase

	position, err := uc.getPositionBySide(ctx, tx, order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:getPositionBySide [%+v] error: %v", order, err))
		return err
	}

	unhold := order.Amount
	if unhold > position.HoldAmount {
		unhold = position.HoldAmount
	}

	position.HoldAmount -= unhold
	if position.HoldAmount < 0 {
		position.HoldAmount = 0
	}

	if order.Side == entities.OrderSideBuy {
		position.Amount += order.Amount
	} else {
		position.Amount -= order.Amount
	}

	if order.Amount > unhold {
		if position.Price == 0 {
			position.Price = order.Price
		} else {
			position.Price = (position.Price + order.Price) / 2
		}
	}

	position.Margin = math.Abs(position.Amount) * position.Price / float64(position.Leverage)
	position.UpdateTS = order.UpdateTS

	err = uc.savePosition(ctx, tx, position)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:savePosition [%+v] error: %v", position, err))
		return err
	}

	if unhold > 0 {
		amount := unhold * order.Price

		wallet := entities.Wallet{
			Exchange:   order.Exchange,
			AccountUID: order.AccountUID,
			Balance: entities.Balance{
				Coin:  coin,
				Total: amount,
			},
			UpdateTS: entities.TS(),
		}

		err = uc.wallet.AppendTotalCoin(ctx, tx, wallet)
		if err != nil {
			uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:AppendTotalCoin [AccountUID: %s, coin: %s, amount: %v] error: %v", order.AccountUID, coin, amount, err))
			return err
		}
	}

	order.Amount -= unhold
	if order.Amount > 0 {
		balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
		if err != nil {
			uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:getBalanceCoin [%+v] error: %v", order, err))
			return err
		}

		cost := order.Amount * order.Price
		hold := balanceHold - cost
		if hold < 0 {
			hold = 0
		}

		wallet := entities.Wallet{
			Exchange:   order.Exchange,
			AccountUID: order.AccountUID,
			Balance: entities.Balance{
				Coin:  coin,
				Total: cost,
				Hold:  hold,
			},
			UpdateTS: entities.TS(),
		}

		err = uc.wallet.SetHoldCoin(ctx, tx, wallet)
		if err != nil {
			uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:SetHoldCoin [AccountUID: %s, coin: %s, hold: %v] error: %v", order.AccountUID, coin, hold, err))
			return err
		}

		if cost > balanceTotal {
			uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:ErrInsufficientFunds [AccountUID: %s, coin: %s, balance_total: %v, cost: %v]", order.AccountUID, coin, balanceTotal, cost))
			return apperror.ErrInsufficientFunds
		}

		err = uc.wallet.SubtractTotalCoin(ctx, tx, wallet)
		if err != nil {
			uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:SubtractTotalCoin [AccountUID: %s, coin: %s, cost: %v] error: %v", order.AccountUID, coin, cost, err))
			return err
		}
	}

	return nil
}
