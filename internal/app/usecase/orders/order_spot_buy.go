package orders

import (
	"context"
	"fmt"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

type OrderSpotBuy struct {
	order *entities.Order
}

func NewOrderSpotBuy(o *entities.Order) *OrderSpotBuy {
	return &OrderSpotBuy{o}
}

func (o *OrderSpotBuy) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
	if err != nil {
		log.Error(fmt.Sprintf("HoldBalance:GetBalanceCoin [%+v] error: %v", o, err))
		return err
	}

	cost := o.order.Amount * o.order.Price
	hold := balanceHold + cost
	if hold > balanceTotal {
		log.Error(fmt.Sprintf("HoldBalance:ErrInsufficientFunds [AccountUID: %s, exchange: %s, coin: %s, balance_total: %v, balance_hold: %v, hold: %v]", o.order.AccountUID, o.order.Exchange, coin, balanceTotal, balanceHold, hold))
		return apperror.ErrInsufficientFunds
	}

	balance := entities.Balance{
		Coin: coin,
		Hold: hold,
	}

	err = uc.SetHoldBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("HoldBalance:SetHoldBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	return nil
}

func (o *OrderSpotBuy) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:GetBalanceCoin [%+v] error: %v", o, err))
		return err
	}

	cost := o.order.Amount * o.order.Price
	hold := balanceHold - cost

	if hold > balanceTotal {
		hold = balanceTotal
	}

	if hold < 0 {
		hold = 0
	}

	balance := entities.Balance{
		Coin: coin,
		Hold: hold,
	}

	err = uc.SetHoldBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:SetHoldBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	return nil
}

func (o *OrderSpotBuy) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.GetBalanceCoin(ctx, o.order.Exchange, o.order.AccountUID, coin)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:GetBalanceCoin [%+v] error: %v", o, err))
		return err
	}

	cost := o.order.Amount * o.order.Price
	if cost > balanceTotal {
		log.Error(fmt.Sprintf("AppendBalance:ErrInsufficientFunds [AccountUID: %s, coin: %s, balance_total: %v, cost: %v]", o.order.AccountUID, coin, balanceTotal, cost))
		return apperror.ErrInsufficientFunds
	}

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

	err = uc.SubtractBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:SubtractBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	appendBalance := entities.Balance{
		Coin:  coins.CoinQuote,
		Total: o.order.Amount,
	}

	err = uc.AppendBalance(ctx, o.order.Exchange, o.order.AccountUID, appendBalance)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:AppendBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, appendBalance, err))
		return err
	}

	return nil
}
