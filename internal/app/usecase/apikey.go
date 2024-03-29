package usecase

import (
	"context"
	"fmt"
	"time"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

func (uc *Usecase) CreateToken(ctx context.Context, service, userID string) (entities.Token, error) {
	var key *entities.Key

	if err := uc.apikey.WithTx(ctx, func(ctx context.Context) error {
		account, err := uc.getAccount(ctx, service, userID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("CreateToken:getAccount [service: %s, user_id: %s] error: %v", service, userID, err))
			return err
		}

		keys, err := uc.apikey.SelectAccountKeys(ctx, account.AccountUID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("CreateToken:SelectAccountKeys [account_uid: %s] error: %v", account.AccountUID, err))
			return err
		}

		if len(keys) >= uc.cfg.KeyLimit {
			return apperror.ErrTokenLimitExceeded
		}

		key = entities.NewToken(account.AccountUID)

		err = uc.apikey.InsertAccountKey(ctx, key)
		if err != nil {
			uc.log.Error(fmt.Sprintf("CreateToken:InsertAccountKey [%+v] error: %v", *key, err))
			return err
		}

		if account.IsNew {
			for _, exchange := range exchanges {
				balance := uc.GetInitialBalance(exchange)

				wallet := entities.Wallet{
					Exchange:   exchange,
					AccountUID: key.AccountUID,
					Balance: entities.Balance{
						Coin:  balance.Coin,
						Total: balance.Total,
					},
					UpdateTS: entities.TS(),
				}

				err := uc.wallet.AppendTotalCoin(ctx, wallet)
				if err != nil {
					uc.log.Error(fmt.Sprintf("CreateToken:AppendTotalCoin [%+v] error: %v", wallet, err))
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return "", err
	}

	uc.log.Info(fmt.Sprintf("CreateToken: [service: %s, user_id: %s]", service, userID))

	return key.Token, nil
}

func (uc *Usecase) GetAccountUID(ctx context.Context, token entities.Token) (entities.AccountUID, error) {
	return uc.apikey.SelectAccountUID(ctx, token)
}

func (uc *Usecase) DisableToken(ctx context.Context, token entities.Token) error {
	key := &entities.Key{
		Token:    token,
		Disabled: true,
		UpdateTS: time.Now().UTC().UnixMilli(),
	}

	err := uc.apikey.UpdateAccountKey(ctx, key)

	uc.log.Info(fmt.Sprintf("DisableToken: [%s]", token))

	return err
}
