package usecase

import (
	"context"
	"errors"
	"fmt"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) GetAccountByUID(ctx context.Context, accountUID entities.AccountUID) (*entities.Account, error) {
	uid := accountUID.CacheUID()

	if account, ok := uc.cacheAccountsByUID.Get(uid); ok {
		return account, nil
	}

	result, err, _ := uc.accountGroupByUID.Do(uid, func() (any, error) {
		if account, ok := uc.cacheAccountsByUID.Get(uid); ok {
			return account, nil
		}

		account, err := uc.account.SelectAccountByUID(ctx, accountUID)
		if err != nil {
			return nil, err
		}

		uc.cacheAccountsByUID.Set(uid, account)

		return account, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.Account), nil
}

func (uc *Usecase) SetAccountPositionMode(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, positionMode entities.PositionMode) error {
	return uc.account.WithTx(ctx, func(ctx context.Context) error {
		if err := uc.checkPresentPendingOrders(ctx, exchange, accountUID, nil); err != nil {
			return apperror.ErrSetPositionMode.Wrap(err)
		}

		if err := uc.checkPresentOpenPosition(ctx, accountUID); err != nil {
			return apperror.ErrSetPositionMode.Wrap(err)
		}

		account := entities.Account{
			AccountUID:   accountUID,
			PositionMode: positionMode,
			UpdateTS:     entities.TS(),
		}

		if err := uc.account.UpdatePositionMode(ctx, &account); err != nil {
			return apperror.ErrSetPositionMode.Wrap(err)
		}

		return nil
	})
}

func (uc *Usecase) AccountClose(ctx context.Context, service, userID string) error {
	return uc.account.WithTx(ctx, func(ctx context.Context) error {
		account, err := uc.account.SelectAccount(ctx, service, userID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("AccountClose:SelectAccount [account_uid: %s] error: %v", account.AccountUID, err))
			return err
		}

		positions, err := uc.position.SelectAccountOpenPositions(ctx, account.AccountUID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("AccountClose:SelectAccountOpenPositions [account_uid: %s] error: %v", account.AccountUID, err))
			return err
		}

		for _, position := range positions {
			position.Amount = 0
			position.HoldAmount = 0
			position.Margin = 0
			position.UpdateTS = entities.TS()

			if err := uc.position.UpdatePosition(ctx, position); err != nil {
				uc.log.Error(fmt.Sprintf("AccountClose:UpdatePosition [%+v] error: %v", *position, err))
				return err
			}

			uc.log.Info(fmt.Sprintf("Close Position: [account_uid: %s, symbol: %s]", account.AccountUID, position.Symbol.String()))
		}

		keys, err := uc.apikey.SelectAccountKeys(ctx, account.AccountUID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("AccountClose:SelectAccountKeys [account_uid: %s] error: %v", account.AccountUID, err))
			return err
		}

		for _, key := range keys {
			if err := uc.DisableToken(ctx, key.Token); err != nil {
				return err
			}
		}

		uid := account.AccountUID.CacheUID()
		uc.cacheAccountsByUID.Delete(uid)

		return uc.disableAccount(ctx, account.AccountUID)
	})
}

func (uc *Usecase) getAccount(ctx context.Context, service, userID string) (*entities.Account, error) {
	account, err := uc.account.SelectAccount(ctx, service, userID)
	if err != nil {
		if errors.Is(err, apperror.ErrAccountNotFound) {
			account, err = uc.createAccount(ctx, service, userID)
			if err != nil {
				return nil, err
			}

			account.IsNew = true
			return account, nil
		}
		return nil, err
	}

	return account, nil
}

func (uc *Usecase) createAccount(ctx context.Context, service, userID string) (*entities.Account, error) {
	account := entities.NewAccount(service, userID)

	err := uc.account.InsertAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	uc.log.Info(fmt.Sprintf("Create Account: [%s]", account.AccountUID))

	return account, nil
}

func (uc *Usecase) disableAccount(ctx context.Context, accountID entities.AccountUID) error {
	account := &entities.Account{
		AccountUID: accountID,
		Disabled:   true,
		UpdateTS:   entities.TS(),
	}

	err := uc.account.UpdateAccount(ctx, account)

	uc.log.Info(fmt.Sprintf("Disable Account: [%s]", account.AccountUID))

	return err
}
