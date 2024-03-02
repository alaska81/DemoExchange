package orders

import (
	"context"
	"fmt"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/pkg/precision"
)

type OrderFuturesHedgeOpen struct {
	order *entities.Order
}

func NewOrderFuturesHedgeOpen(o *entities.Order) *OrderFuturesHedgeOpen {
	return &OrderFuturesHedgeOpen{o}
}

func (o *OrderFuturesHedgeOpen) Validate() error {
	amount := precision.ToFix(o.order.Amount, o.order.Precision)

	if amount <= 0 {
		return apperror.ErrAmountIsOutOfRange
	}

	if o.order.Limit > 0 && amount < o.order.Limit {
		return apperror.ErrAmountIsOutOfRange
	}

	o.order.Amount = amount

	return nil
}

func (o *OrderFuturesHedgeOpen) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
	if err != nil {
		log.Error(fmt.Sprintf("HoldBalance:GetBalanceCoin [%+v] error: %v", o, err))
		return err
	}

	cost := o.order.Amount * o.order.Price

	o.order.Fee = cost * OrderFuturesFee
	o.order.FeeCoin = coin

	cost += o.order.Fee

	hold := balanceHold + cost
	if hold > balanceTotal {
		log.Error(fmt.Sprintf("HoldBalance:ErrInsufficientFunds [AccountUID: %s, exchange: %s, coin: %s, balance_total: %v, balance_hold: %v, cost: %v]", o.order.AccountUID, o.order.Exchange, coin, balanceTotal, balanceHold, cost))
		return apperror.ErrInsufficientFunds
	}

	balance := entities.Balance{
		Coin: coin,
		Hold: hold,
	}

	err = uc.SetHoldBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:SetHoldBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	return nil
}

func (o *OrderFuturesHedgeOpen) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:GetBalanceCoin [%+v] error: %v", o, err))
		return err
	}

	cost := o.order.Amount * o.order.Price

	o.order.Fee = cost * OrderFuturesFee
	o.order.FeeCoin = coin

	cost += o.order.Fee
	unhold := balanceHold - cost
	if unhold > balanceTotal {
		unhold = balanceTotal
	}

	if unhold < 0 {
		unhold = 0
	}

	balance := entities.Balance{
		Coin: coin,
		Hold: unhold,
	}

	err = uc.SetHoldBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:SetHoldBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	return nil
}

func (o *OrderFuturesHedgeOpen) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:GetBalanceCoin [%+v] error: %v", o, err))
		return err
	}

	cost := o.order.Amount * o.order.Price

	o.order.Fee = cost * OrderFuturesFee
	o.order.FeeCoin = coin

	cost += o.order.Fee
	if cost > balanceTotal {
		log.Error(fmt.Sprintf("AppendBalance:ErrInsufficientFunds [AccountUID: %s, coin: %s, balance_total: %v, cost: %v]", o.order.AccountUID, coin, balanceTotal, cost))
		return apperror.ErrInsufficientFunds
	}

	unhold := balanceHold - cost
	if unhold < 0 {
		unhold = 0
	}

	balance := entities.Balance{
		Coin:  coin,
		Total: cost,
		Hold:  unhold,
	}

	err = uc.SetHoldBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:SetHoldBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	err = uc.SubtractBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:SubtractBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	position, err := uc.GetPositionBySide(ctx, o.order.Exchange, o.order.AccountUID, o.order.Symbol, o.order.PositionSide)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:GetPositionBySide [%+v] error: %v", o, err))
		return err
	}

	if position.Amount == 0 {
		position.Price = o.order.Price
	} else {
		position.Price = (position.Price + o.order.Price) / 2
	}

	position.Amount += o.order.Amount
	position.Margin = position.Amount * position.Price / float64(position.Leverage)

	position.UpdateTS = o.order.UpdateTS

	err = uc.SavePosition(ctx, position)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:SavePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}
