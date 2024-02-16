package usecase

import (
	"context"
	"errors"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) createAccount(ctx context.Context, service, userID string) (*entities.Account, error) {
	account := entities.NewAccount(service, userID)

	err := uc.account.InsertAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	return account, nil
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

func (uc *Usecase) GetAccountByUID(ctx context.Context, accountUID entities.AccountUID) (*entities.Account, error) {
	return uc.account.SelectAccountByUID(ctx, accountUID)
}

func (uc *Usecase) SetAccountPositionMode(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, positionMode entities.PositionMode) error {
	return uc.account.WithTx(ctx, func(ctx context.Context) error {
		if err := uc.checkPresentPendingOrders(ctx, exchange, accountUID, nil); err != nil {
			return apperror.ErrSetPositionMode.Wrap(err)
		}

		if err := uc.checkPresentOpenPosition(ctx, exchange, accountUID); err != nil {
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
