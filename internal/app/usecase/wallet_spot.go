package usecase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) holdBalanceSpot(ctx context.Context, order *entities.Order) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if order.Side == entities.OrderSideBuy {
			return uc.holdBalanceSpotBuy(ctx, tx, order)
		} else {
			return uc.holdBalanceSpotSell(ctx, tx, order)
		}
	})
}

func (uc *Usecase) holdBalanceSpotBuy(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		return err
	}

	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceSpotBuy:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	cost := order.Amount * order.Price
	hold := balanceHold + cost
	if hold > balanceTotal {
		uc.log.Error(fmt.Sprintf("holdBalanceSpotBuy:ErrInsufficientFunds [AccountUID: %s, exchange: %s, coin: %s, balance_total: %v, balance_hold: %v, hold: %v]", order.AccountUID, order.Exchange, coin, balanceTotal, balanceHold, hold))
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
		uc.log.Error(fmt.Sprintf("holdBalanceSpotBuy:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, hold: %v] error: %v", order.AccountUID, order.Exchange, coin, hold, err))
		return err
	}

	return nil
}

func (uc *Usecase) holdBalanceSpotSell(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		return err
	}

	coin := coins.CoinQuote

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("holdBalanceSpotSell:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	hold := balanceHold + order.Amount
	if balanceTotal < hold {
		uc.log.Error(fmt.Sprintf("holdBalanceSpotSell:ErrInsufficientFunds [AccountUID: %s, exchange: %s, coin: %s, balance_total: %v, balance_hold: %v, hold: %v]", order.AccountUID, order.Exchange, coin, balanceTotal, balanceHold, hold))
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
		uc.log.Error(fmt.Sprintf("holdBalanceSpotSell:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, hold: %v] error: %v", order.AccountUID, order.Exchange, coin, hold, err))
		return err
	}

	return nil
}

func (uc *Usecase) unholdBalanceSpot(ctx context.Context, order *entities.Order) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if order.Side == entities.OrderSideBuy {
			return uc.unholdBalanceSpotBuy(ctx, tx, order)
		} else {
			return uc.unholdBalanceSpotSell(ctx, tx, order)
		}
	})
}

func (uc *Usecase) unholdBalanceSpotBuy(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		return err
	}

	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceSpotBuy:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	cost := order.Amount * order.Price
	hold := balanceHold - cost

	if hold > balanceTotal {
		hold = balanceTotal
	}

	if hold < 0 {
		hold = 0
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
		uc.log.Error(fmt.Sprintf("unholdBalanceSpotBuy:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, hold: %v] error: %v", order.AccountUID, order.Exchange, coin, hold, err))
		return err
	}

	return nil
}

func (uc *Usecase) unholdBalanceSpotSell(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		return err
	}

	coin := coins.CoinQuote

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("unholdBalanceSpotSell:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	hold := balanceHold - order.Amount

	if hold > balanceTotal {
		hold = balanceTotal
	}

	if hold < 0 {
		hold = 0
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
		uc.log.Error(fmt.Sprintf("unholdBalanceSpotSell:SetHoldCoin [AccountUID: %s, exchange: %s, coin: %s, hold: %v] error: %v", order.AccountUID, order.Exchange, coin, hold, err))
		return err
	}

	return nil
}

func (uc *Usecase) appendBalanceSpot(ctx context.Context, order *entities.Order) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		if order.Side == entities.OrderSideBuy {
			return uc.appendBalanceSpotBuy(ctx, tx, order)
		} else {
			return uc.appendBalanceSpotSell(ctx, tx, order)
		}
	})
}

func (uc *Usecase) appendBalanceSpotBuy(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotBuy:GetCoins [%+v] error: %v", order, err))
		return err
	}

	coin := coins.CoinBase

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotBuy:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	cost := order.Amount * order.Price
	if cost > balanceTotal {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotBuy:ErrInsufficientFunds [AccountUID: %s, coin: %s, balance_total: %v, cost: %v]", order.AccountUID, coin, balanceTotal, cost))
		return apperror.ErrInsufficientFunds
	}

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
		uc.log.Error(fmt.Sprintf("appendBalanceSpotBuy:SetHoldCoin [AccountUID: %s, coin: %s, hold: %v] error: %v", order.AccountUID, coin, hold, err))
		return err
	}

	err = uc.wallet.SubtractTotalCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotBuy:SubtractTotalCoin [AccountUID: %s, coin: %s, cost: %v] error: %v", order.AccountUID, coin, cost, err))
		return err
	}

	wallet.Balance.Coin = coins.CoinQuote
	wallet.Balance.Total = order.Amount

	err = uc.wallet.AppendTotalCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotBuy:AppendTotalCoin [AccountUID: %s, coin: %s, amount: %v] error: %v", order.AccountUID, coins.CoinQuote, order.Amount, err))
		return err
	}

	return nil
}

func (uc *Usecase) appendBalanceSpotSell(ctx context.Context, tx pgx.Tx, order *entities.Order) error {
	coins, err := order.Symbol.GetCoins()
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotSell:GetCoins [%+v] error: %v", order, err))
		return err
	}

	coin := coins.CoinQuote

	balanceTotal, balanceHold, err := uc.getBalanceCoin(ctx, tx, order.Exchange, order.AccountUID, coin)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotSell:getBalanceCoin [%+v] error: %v", order, err))
		return err
	}

	if order.Amount > balanceTotal {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotSell:ErrInsufficientFunds [AccountUID: %s, coin: %s, balance_total: %v, amount: %v]", order.AccountUID, coin, balanceTotal, order.Amount))
		return apperror.ErrInsufficientFunds
	}

	hold := balanceHold - order.Amount
	if hold < 0 {
		hold = 0
	}

	wallet := entities.Wallet{
		Exchange:   order.Exchange,
		AccountUID: order.AccountUID,
		Balance: entities.Balance{
			Coin:  coin,
			Total: order.Amount,
			Hold:  hold,
		},
		UpdateTS: entities.TS(),
	}

	err = uc.wallet.SetHoldCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotSell:SetHoldCoin [AccountUID: %s, coin: %s, hold: %v] error: %v", order.AccountUID, coin, hold, err))
		return err
	}

	err = uc.wallet.SubtractTotalCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotSell:SubtractTotalCoin [AccountUID: %s, coin: %s, amount: %v] error: %v", order.AccountUID, coin, order.Amount, err))
		return err
	}

	cost := order.Amount * order.Price
	wallet.Balance.Coin = coins.CoinBase
	wallet.Balance.Total = cost
	err = uc.wallet.AppendTotalCoin(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("appendBalanceSpotSell:AppendTotalCoin [AccountUID: %s, coin: %s, order_cost: %v] error: %v", order.AccountUID, coins.CoinBase, cost, err))
		return err
	}

	return nil
}
