package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

var muWallet = &sync.RWMutex{}

func (uc *Usecase) GetBalances(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) (entities.Balances, error) {
	muWallet.RLock()
	defer muWallet.RUnlock()

	wallet := entities.Wallet{
		Exchange:   exchange,
		AccountUID: accountUID,
	}
	return uc.wallet.SelectBalances(ctx, nil, wallet)
}

func (uc *Usecase) Deposit(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) (float64, error) {
	return uc.AppendCoin(ctx, exchange, accountUID, coin, amount)
}

func (uc *Usecase) AppendCoin(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) (float64, error) {
	muWallet.Lock()
	defer muWallet.Unlock()

	err := uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		wallet := entities.Wallet{
			Exchange:   exchange,
			AccountUID: accountUID,
		}
		balances, err := uc.wallet.SelectBalances(ctx, tx, wallet)
		if err != nil {
			return err
		}

		if balances[coin].Total+amount > uc.getBalanceLimit(coin) {
			amount = uc.getBalanceLimit(coin) - balances[coin].Total
		}

		if amount <= 0 {
			return apperror.ErrBalanceLimitExceeded
		}

		wallet.Balance.Coin = coin
		wallet.Balance.Total = amount
		wallet.UpdateTS = entities.TS()

		return uc.wallet.AppendTotalCoin(ctx, tx, wallet)
	})

	return amount, err
}

func (uc *Usecase) Withdraw(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) error {
	return uc.SubtractCoin(ctx, exchange, accountUID, coin, amount)
}

func (uc *Usecase) SubtractCoin(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
		wallet := entities.Wallet{
			Exchange:   exchange,
			AccountUID: accountUID,
		}
		balances, err := uc.wallet.SelectBalances(ctx, tx, wallet)
		if err != nil {
			return err
		}

		if balances[coin].Total-amount < 0 {
			return apperror.ErrInsufficientFunds
		}

		wallet.Balance.Coin = coin
		wallet.Balance.Total = amount
		wallet.UpdateTS = entities.TS()

		return uc.wallet.SubtractTotalCoin(ctx, tx, wallet)
	})
}

func (uc *Usecase) getBalanceCoin(ctx context.Context, tx pgx.Tx, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin) (total float64, hold float64, err error) {
	var balances entities.Balances

	wallet := entities.Wallet{
		Exchange:   exchange,
		AccountUID: accountUID,
	}
	balances, err = uc.wallet.SelectBalances(ctx, tx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("getBalanceCoin:SelectBalances [%+v] error: %v", wallet, err))
		return
	}

	total = balances[coin].Total
	hold = balances[coin].Hold

	return
}

func (uc *Usecase) holdBalance(ctx context.Context, order *entities.Order) error {
	if order.Exchange == entities.ExchangeSpot {
		return uc.holdBalanceSpot(ctx, order)
	} else {
		return uc.holdBalanceFutures(ctx, order)
	}
}

func (uc *Usecase) unholdBalance(ctx context.Context, order *entities.Order) error {
	if order.Exchange == entities.ExchangeSpot {
		return uc.unholdBalanceSpot(ctx, order)
	} else {
		return uc.unholdBalanceFutures(ctx, order)
	}
}

func (uc *Usecase) appendBalance(ctx context.Context, order *entities.Order) error {
	if order.Exchange == entities.ExchangeSpot {
		return uc.appendBalanceSpot(ctx, order)
	} else {
		return uc.appendBalanceFutures(ctx, order)
	}
}

func (uc *Usecase) getBalanceLimit(coin entities.Coin) float64 {
	return uc.cfg.MaxBalances[coin]
}
