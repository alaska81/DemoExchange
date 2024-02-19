package usecase

import (
	"context"
	"fmt"
	"sync"

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
	balances, err := uc.wallet.SelectBalances(ctx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("GetBalances:SelectBalances [%+v] error: %v", wallet, err))
		return nil, err
	}

	if exchange == entities.ExchangeSpot {
		return balances, nil
	}

	for coin, balance := range balances {
		available := balance.Total - balance.Hold
		balance.AvailableBalance = available
		balance.WalletBalance = available
		balances[coin] = balance
	}

	positions, err := uc.PositionsList(ctx, exchange, accountUID)
	if err != nil {
		return nil, err
	}

	for _, p := range positions {
		coins := p.Symbol.GetCoins()

		if balance, ok := balances[coins.CoinBase]; ok {
			balance.InitialMargin += p.Margin
			balance.WalletBalance += p.Margin        // available + margin
			balance.MarginBalance += p.MarginBalance // margin + pnl
			balance.UnrealisedPnl += p.UnrealisedPnl
			balances[coins.CoinBase] = balance
		}
	}

	return balances, nil
}

func (uc *Usecase) Deposit(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) (float64, error) {
	return uc.AppendCoin(ctx, exchange, accountUID, coin, amount)
}

func (uc *Usecase) AppendCoin(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) (float64, error) {
	muWallet.Lock()
	defer muWallet.Unlock()

	err := uc.wallet.WithTx(ctx, func(ctx context.Context) error {
		wallet := entities.Wallet{
			Exchange:   exchange,
			AccountUID: accountUID,
		}
		balances, err := uc.wallet.SelectBalances(ctx, wallet)
		if err != nil {
			return err
		}

		balance, ok := balances[coin]
		if !ok {
			return apperror.ErrBalanceNotFound
		}

		if balance.Total+amount > uc.getBalanceLimit(coin) {
			amount = uc.getBalanceLimit(coin) - balance.Total
		}

		if amount <= 0 {
			return apperror.ErrBalanceLimitExceeded
		}

		wallet.Balance.Coin = coin
		wallet.Balance.Total = amount
		wallet.UpdateTS = entities.TS()

		return uc.wallet.AppendTotalCoin(ctx, wallet)
	})

	return amount, err
}

func (uc *Usecase) Withdraw(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) error {
	return uc.SubtractCoin(ctx, exchange, accountUID, coin, amount)
}

func (uc *Usecase) SubtractCoin(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin, amount float64) error {
	muWallet.Lock()
	defer muWallet.Unlock()

	return uc.wallet.WithTx(ctx, func(ctx context.Context) error {
		wallet := entities.Wallet{
			Exchange:   exchange,
			AccountUID: accountUID,
		}
		balances, err := uc.wallet.SelectBalances(ctx, wallet)
		if err != nil {
			return err
		}

		balance, ok := balances[coin]
		if !ok {
			return apperror.ErrBalanceNotFound
		}

		if balance.Total-amount < 0 {
			return apperror.ErrInsufficientFunds
		}

		wallet.Balance.Coin = coin
		wallet.Balance.Total = amount
		wallet.UpdateTS = entities.TS()

		return uc.wallet.SubtractTotalCoin(ctx, wallet)
	})
}

func (uc *Usecase) GetBalanceCoin(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, coin entities.Coin) (total float64, hold float64, err error) {
	var balances entities.Balances

	wallet := entities.Wallet{
		Exchange:   exchange,
		AccountUID: accountUID,
	}
	balances, err = uc.wallet.SelectBalances(ctx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("GetBalanceCoin:SelectBalances [%+v] error: %v", wallet, err))
		return
	}

	balance, ok := balances[coin]
	if !ok {
		err = apperror.ErrBalanceNotFound
		return
	}

	total = balance.Total
	hold = balance.Hold

	return
}

func (uc *Usecase) SetHoldBalance(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, balance entities.Balance) error {
	wallet := entities.Wallet{
		Exchange:   exchange,
		AccountUID: accountUID,
		Balance:    balance,
		UpdateTS:   entities.TS(),
	}
	err := uc.wallet.SetHoldCoin(ctx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("SetHoldBalance:SetHoldCoin [%+v] error: %v", wallet, err))
		return err
	}

	return nil
}

func (uc *Usecase) SubtractBalance(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, balance entities.Balance) error {
	wallet := entities.Wallet{
		Exchange:   exchange,
		AccountUID: accountUID,
		Balance:    balance,
		UpdateTS:   entities.TS(),
	}
	err := uc.wallet.SubtractTotalCoin(ctx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("SubtractBalance:SubtractTotalCoin [%+v] error: %v", wallet, err))
		return err
	}

	return nil
}

func (uc *Usecase) AppendBalance(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, balance entities.Balance) error {
	wallet := entities.Wallet{
		Exchange:   exchange,
		AccountUID: accountUID,
		Balance:    balance,
		UpdateTS:   entities.TS(),
	}
	err := uc.wallet.AppendTotalCoin(ctx, wallet)
	if err != nil {
		uc.log.Error(fmt.Sprintf("AppendBalance:AppendTotalCoin [%+v] error: %v", wallet, err))
		return err
	}

	return nil
}

func (uc *Usecase) getBalanceLimit(coin entities.Coin) float64 {
	return uc.cfg.MaxBalances[coin]
}
