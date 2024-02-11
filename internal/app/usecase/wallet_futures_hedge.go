package usecase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) holdBalanceFuturesHedge(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	if order.Side == entities.OrderSideBuy {
		return uc.holdBalanceFuturesHedgeOpen(ctx, tx, order)
	} else {
		return uc.holdBalanceFuturesHedgeClose(ctx, tx, order)
	}
}

func (uc *Usecase) holdBalanceFuturesHedgeOpen(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeOpen:GetCoins [%+v] error: %v", order, err))
		return err
	}

	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeOpen:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	cost := order.Amount * order.Price
	hold := balanceHold + cost
	if hold > balanceTotal {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeOpen:ErrInsufficientFunds [AccountUID: %s, exchange: %s, coin: %s, balance_total: %v, balance_hold: %v, cost: %v]", order.AccountUID, order.Exchange, coin, balanceTotal, balanceHold, cost))
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
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeOpen:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, hold: %v] error: %v", order.AccountUID, order.Exchange, coin, hold, err))
		return err
	}

	return nil
}

func (uc *Usecase) holdBalanceFuturesHedgeClose(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	position, err := uc.getPositionBySide(ctx, tx, order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeClose:getPositionBySide [%+v] error: %v", order, err))
		return err
	}

	if position.IsNew {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeClose:ErrPositionNotFound [AccountUID: %s, exchange: %s, symbol: %s, position_side: %s]", order.AccountUID, order.Exchange, order.Symbol, order.PositionSide))
		return apperror.ErrPositionNotFound
	}

	positionBalance := position.Amount - position.HoldAmount

	if order.Amount > positionBalance {
		order.Amount = positionBalance
	}

	if order.Amount == 0 {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeClose:ErrInsufficientFunds [AccountUID: %s, position_amount: %v, order_amount: %v]", order.AccountUID, position.Amount, order.Amount))
		return apperror.ErrInsufficientFunds
	}

	position.HoldAmount += order.Amount
	position.UpdateTS = order.UpdateTS

	err = uc.position.UpdatePosition(ctx, tx, position)
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeClose:UpdatePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}

func (uc *Usecase) unholdBalanceFuturesHedge(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	if order.Side == entities.OrderSideBuy {
		return uc.unholdBalanceFuturesHedgeOpen(ctx, tx, order)
	} else {
		return uc.unholdBalanceFuturesHedgeClose(ctx, tx, order)
	}
}

func (uc *Usecase) unholdBalanceFuturesHedgeOpen(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeOpen:GetCoins [%+v] error: %v", order, err))
		return err
	}

	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceFuturesHedgeOpen:getBalanceCoin [%+v] error: %v", order, err))
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
		uc.log.Error(fmt.Sprintf("unholdBalanceFuturesHedgeOpen:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, unhold: %v] error: %v", order.AccountUID, order.Exchange, coin, unhold, err))
		return err
	}

	return nil
}

func (uc *Usecase) unholdBalanceFuturesHedgeClose(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	position, err := uc.getPositionBySide(ctx, tx, order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceFuturesHedgeClose:getPositionBySide [%+v] error: %v", order, err))
		return err
	}

	position.HoldAmount -= order.Amount
	if position.HoldAmount < 0 {
		position.HoldAmount = 0
	}

	err = uc.position.UpdatePosition(ctx, tx, position)
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceFuturesHedgeClose:UpdatePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}

func (uc *Usecase) appendBalanceFuturesHedge(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	if order.Side == entities.OrderSideBuy {
		return uc.appendBalanceFuturesHedgeOpen(ctx, tx, order)
	} else {
		return uc.appendBalanceFuturesHedgeClose(ctx, tx, order)
	}
}

func (uc *Usecase) appendBalanceFuturesHedgeOpen(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:GetCoins [%+v] error: %v", order, err))
		return err
	}

	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	cost := order.Amount * order.Price
	if cost > balanceTotal {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:ErrInsufficientFunds [AccountUID: %s, coin: %s, balance_total: %v, cost: %v]", order.AccountUID, coin, balanceTotal, cost))
		return apperror.ErrInsufficientFunds
	}

	unhold := balanceHold - cost
	if unhold < 0 {
		unhold = 0
	}

	wallet := entities.Wallet{
		Exchange:   order.Exchange,
		AccountUID: order.AccountUID,
		Balance: entities.Balance{
			Coin:  coin,
			Total: cost,
			Hold:  unhold,
		},
		UpdateTS: entities.TS(),
	}

	err = uc.wallet.SetHoldCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:SetHoldCoin [AccountUID: %s, coin: %s, unhold: %v] error: %v", order.AccountUID, coin, unhold, err))
		return err
	}

	err = uc.wallet.SubtractTotalCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:SubtractTotalCoin [AccountUID: %s, coin: %s, cost: %v] error: %v", order.AccountUID, coin, cost, err))
		return err
	}

	position, err := uc.getPositionBySide(ctx, tx, order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:getPositionBySide [%+v] error: %v", order, err))
		return err
	}

	position.Amount += order.Amount

	if position.Price == 0 {
		position.Price = order.Price
	} else {
		position.Price = (position.Price + order.Price) / 2
	}

	position.UpdateTS = order.UpdateTS

	err = uc.savePosition(ctx, tx, position)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:savePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}

func (uc *Usecase) appendBalanceFuturesHedgeClose(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeOpen:GetCoins [%+v] error: %v", order, err))
		return err
	}

	coin := coins.CoinBase

	position, err := uc.getPositionBySide(ctx, tx, order)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeClose:getPositionBySide [%+v] error: %v", order, err))
		return err
	}

	position.HoldAmount -= order.Amount
	if position.HoldAmount < 0 {
		position.HoldAmount = 0
	}

	position.Amount -= order.Amount
	if position.Amount < 0 {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeClose:ErrInsufficientFunds [AccountUID: %s, position_amount: %v, order_amount: %v]", order.AccountUID, position.Amount, order.Amount))
		return apperror.ErrInsufficientFunds
	}

	position.UpdateTS = order.UpdateTS

	err = uc.savePosition(ctx, tx, position)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesHedgeClose:savePosition [%+v] error: %v", position, err))
		return err
	}

	cost := order.Amount * order.Price

	wallet := entities.Wallet{
		Exchange:   order.Exchange,
		AccountUID: order.AccountUID,
		Balance: entities.Balance{
			Coin:  coin,
			Total: cost,
		},
		UpdateTS: entities.TS(),
	}

	err = uc.wallet.AppendTotalCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceFuturesOneway:AppendTotalCoin [AccountUID: %s, coin: %s, cost: %v] error: %v", order.AccountUID, coin, cost, err))
		return err
	}

	return nil
}
