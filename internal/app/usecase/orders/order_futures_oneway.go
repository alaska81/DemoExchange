package orders

import (
	"context"
	"fmt"
	"math"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

type OrderFuturesOneway struct {
	order *entities.Order
}

func NewOrderFuturesOneway(order *entities.Order) *OrderFuturesOneway {
	return &OrderFuturesOneway{order}
}

func (o *OrderFuturesOneway) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	position, err := uc.GetPositionBySide(ctx, o.order.Exchange, o.order.AccountUID, o.order.Symbol, o.order.PositionSide)
	if err != nil {
		log.Error(fmt.Sprintf("HoldBalance:GetPositionBySide [%+v] error: %v", o, err))
		return err
	}

	balancePosition := 0.0
	if o.order.Side == entities.OrderSideBuy && position.Amount < 0 {
		balancePosition = -position.Amount - position.HoldAmount
	} else if o.order.Side == entities.OrderSideSell && position.Amount > 0 {
		balancePosition = position.Amount - position.HoldAmount
	}

	holdPosition := o.order.Amount
	if holdPosition > balancePosition {
		holdPosition = balancePosition
	}

	if holdPosition > 0 {
		position.HoldAmount += holdPosition
		position.UpdateTS = o.order.UpdateTS

		err = uc.SavePosition(ctx, position)
		if err != nil {
			log.Error(fmt.Sprintf("HoldBalance:SavePosition [%+v] error: %v", position, err))
			return err
		}
	}

	hold := (o.order.Amount - holdPosition) * o.order.Price
	if hold > 0 {
		coins := o.order.Symbol.GetCoins()
		coin := coins.CoinBase

		balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
		if err != nil {
			log.Error(fmt.Sprintf("HoldBalance:GetBalanceCoin [%+v] error: %v", o, err))
			return err
		}

		if hold > balanceTotal-balanceHold {
			log.Error(fmt.Sprintf("HoldBalance:ErrInsufficientFunds [AccountUID: %s, exchange: %s, coin: %s, balance_total: %v, balance_hold: %v, hold: %v]", o.order.AccountUID, o.order.Exchange, coin, balanceTotal, balanceHold, hold))
			return apperror.ErrInsufficientFunds
		}

		balance := entities.Balance{
			Coin: coin,
			Hold: hold + balanceHold,
		}

		err = uc.SetHoldBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
		if err != nil {
			log.Error(fmt.Sprintf("HoldBalance:SetHoldBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
			return err
		}
	}

	return nil
}

func (o *OrderFuturesOneway) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:GetBalanceCoin [%+v] error: %v", o, err))
		return err
	}

	cost := o.order.Amount * o.order.Price
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

	position, err := uc.GetPositionBySide(ctx, o.order.Exchange, o.order.AccountUID, o.order.Symbol, o.order.PositionSide)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:GetPositionBySide [%+v] error: %v", o, err))
		return err
	}

	position.HoldAmount -= (o.order.Amount - (balanceHold+unhold)/o.order.Price)

	if position.HoldAmount > position.Amount {
		position.HoldAmount = position.Amount
	}

	if position.HoldAmount < 0 {
		position.HoldAmount = 0
	}

	err = uc.SavePosition(ctx, position)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:SavePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}

func (o *OrderFuturesOneway) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	position, err := uc.GetPositionBySide(ctx, o.order.Exchange, o.order.AccountUID, o.order.Symbol, o.order.PositionSide)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:GetPositionBySide [%+v] error: %v", o, err))
		return err
	}

	unhold := o.order.Amount
	if unhold > position.HoldAmount {
		unhold = position.HoldAmount
	}

	position.HoldAmount -= unhold
	if position.HoldAmount < 0 {
		position.HoldAmount = 0
	}

	if o.order.Amount > unhold {
		if position.Amount == 0 {
			position.Price = o.order.Price
		} else {
			position.Price = (position.Price + o.order.Price) / 2
		}
	}

	if o.order.Side == entities.OrderSideBuy {
		position.Amount += o.order.Amount
	} else {
		position.Amount -= o.order.Amount
	}

	position.Margin = math.Abs(position.Amount) * position.Price / float64(position.Leverage)
	position.UpdateTS = o.order.UpdateTS

	err = uc.SavePosition(ctx, position)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:SavePosition [%+v] error: %v", position, err))
		return err
	}

	if unhold > 0 {
		amount := unhold * o.order.Price

		balance := entities.Balance{
			Coin:  coin,
			Total: amount,
		}

		err = uc.AppendBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
		if err != nil {
			log.Error(fmt.Sprintf("AppendBalance:AppendBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
			return err
		}
	}

	o.order.Amount -= unhold
	if o.order.Amount > 0 {
		balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
		if err != nil {
			log.Error(fmt.Sprintf("AppendBalance:GetBalanceCoin [%+v] error: %v", o, err))
			return err
		}

		cost := o.order.Amount * o.order.Price
		hold := balanceHold - cost
		if hold < 0 {
			hold = 0
		}

		balance := entities.Balance{
			Coin:  coin,
			Total: cost,
			Hold:  hold,
		}

		err = uc.SetHoldBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
		if err != nil {
			log.Error(fmt.Sprintf("AppendBalance:SetHoldBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
			return err
		}

		if cost > balanceTotal {
			log.Error(fmt.Sprintf("AppendBalance:ErrInsufficientFunds [AccountUID: %s, coin: %s, balance_total: %v, cost: %v]", o.order.AccountUID, coin, balanceTotal, cost))
			return apperror.ErrInsufficientFunds
		}

		err = uc.SubtractBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
		if err != nil {
			log.Error(fmt.Sprintf("AppendBalance:SubtractBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
			return err
		}
	}

	return nil
}
