package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) createAccount(ctx context.Context, tx pgx.Tx, service, userID string) (*entities.Account, error) {
	account := entities.NewAccount(service, userID)

	err := uc.account.InsertAccount(ctx, tx, account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (uc *Usecase) getAccount(ctx context.Context, tx pgx.Tx, service, userID string) (*entities.Account, error) {
	account, err := uc.account.SelectAccount(ctx, tx, service, userID)
	if err != nil {
		if errors.Is(err, apperror.ErrAccountNotFound) {
			account, err = uc.createAccount(ctx, tx, service, userID)
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

func (uc *Usecase) getAccountByUID(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID) (*entities.Account, error) {
	return uc.account.SelectAccountByUID(ctx, tx, accountUID)
}

func (uc *Usecase) SetAccountPositionMode(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, positionMode entities.PositionMode) error {
	return uc.tx.WithTX(ctx, func(tx pgx.Tx) error {
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
