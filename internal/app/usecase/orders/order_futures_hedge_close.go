package orders

import (
	"context"
	"fmt"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

type OrderFuturesHedgeClose struct {
	order *entities.Order
}

func NewOrderFuturesHedgeClose(o *entities.Order) *OrderFuturesHedgeClose {
	return &OrderFuturesHedgeClose{o}
}

func (o *OrderFuturesHedgeClose) HoldBalance(ctx context.Context, uc Usecase, log Logger) error {
	position, err := uc.GetPositionBySide(ctx, o.order.Exchange, o.order.AccountUID, o.order.Symbol, o.order.PositionSide)
	if err != nil {
		log.Error(fmt.Sprintf("HoldBalance:GetPositionBySide [%+v] error: %v", o, err))
		return err
	}

	if position.IsNew {
		log.Error(fmt.Sprintf("HoldBalance:ErrPositionNotFound [AccountUID: %s, exchange: %s, symbol: %s, position_side: %s]", o.order.AccountUID, o.order.Exchange, o.order.Symbol, o.order.PositionSide))
		return apperror.ErrPositionNotFound
	}

	positionBalance := position.Amount - position.HoldAmount

	if o.order.Amount > positionBalance {
		o.order.Amount = positionBalance
	}

	if o.order.Amount == 0 {
		log.Error(fmt.Sprintf("HoldBalance:ErrInsufficientFunds [AccountUID: %s, position_amount: %v, order_amount: %v]", o.order.AccountUID, position.Amount, o.order.Amount))
		return apperror.ErrInsufficientFunds
	}

	position.HoldAmount += o.order.Amount
	position.UpdateTS = o.order.UpdateTS

	err = uc.SavePosition(ctx, position)
	if err != nil {
		log.Error(fmt.Sprintf("HoldBalance:SavePosition [%+v] error: %v", position, err))
		return err
	}

	return nil
}

func (o *OrderFuturesHedgeClose) UnholdBalance(ctx context.Context, uc Usecase, log Logger) error {
	position, err := uc.GetPositionBySide(ctx, o.order.Exchange, o.order.AccountUID, o.order.Symbol, o.order.PositionSide)
	if err != nil {
		log.Error(fmt.Sprintf("UnholdBalance:GetPositionBySide [%+v] error: %v", o, err))
		return err
	}

	position.HoldAmount -= o.order.Amount
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

func (o *OrderFuturesHedgeClose) AppendBalance(ctx context.Context, uc Usecase, log Logger) error {
	coins := o.order.Symbol.GetCoins()
	coin := coins.CoinBase

	position, err := uc.GetPositionBySide(ctx, o.order.Exchange, o.order.AccountUID, o.order.Symbol, o.order.PositionSide)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:GetPositionBySide [%+v] error: %v", o, err))
		return err
	}

	position.HoldAmount -= o.order.Amount
	if position.HoldAmount < 0 {
		position.HoldAmount = 0
	}

	position.Amount -= o.order.Amount
	if position.Amount < 0 {
		log.Error(fmt.Sprintf("AppendBalance:ErrInsufficientFunds [AccountUID: %s, position_amount: %v, order_amount: %v]", o.order.AccountUID, position.Amount, o.order.Amount))
		return apperror.ErrInsufficientFunds
	}

	position.Margin = position.Amount * position.Price / float64(position.Leverage)
	position.UpdateTS = o.order.UpdateTS

	err = uc.SavePosition(ctx, position)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:SavePosition [%+v] error: %v", position, err))
		return err
	}

	cost := o.order.Amount * o.order.Price

	o.order.Fee = cost * OrderFuturesFee
	o.order.FeeCoin = coin

	balance := entities.Balance{
		Coin:  coin,
		Total: cost - o.order.Fee,
	}

	err = uc.AppendBalance(ctx, o.order.Exchange, o.order.AccountUID, balance)
	if err != nil {
		log.Error(fmt.Sprintf("AppendBalance:AppendBalance [AccountUID: %s, exchange: %s, balance: %+v] error: %v", o.order.AccountUID, o.order.Exchange, balance, err))
		return err
	}

	return nil
}
