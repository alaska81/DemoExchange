package usecase

import (
	"context"
	"fmt"
	"time"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

var exchanges = []entities.Exchange{entities.ExchangeSpot, entities.ExchangeFutures}

var initialBalance = map[entities.Exchange]entities.Balance{
	entities.ExchangeSpot: {
		Coin:  "USDT",
		Total: 3000,
	},
	entities.ExchangeFutures: {
		Coin:  "USDT",
		Total: 3000,
	},
}

func (uc *Usecase) CreateToken(ctx context.Context, service, userID string) (entities.Token, error) {
	var key *entities.Key

	err := uc.apikey.WithTx(ctx, func(ctx context.Context) error {
		account, err := uc.getAccount(ctx, service, userID)
		if err != nil {
			uc.log.Error(fmt.Sprintf("CreateToken:getAccount [service: %s, user_id: %s] error: %v", service, userID, err))
			return err
		}

		keys, err := uc.apikey.SelectAccountKeys(ctx, account.AccountUID)
		if err != nil {
			return err
		}

		if len(keys) >= uc.cfg.KeyLimit {
			return apperror.ErrTokenLimitExceeded
		}

		key = entities.NewToken(account.AccountUID)

		err = uc.apikey.InsertAccountKey(ctx, key)
		if err != nil {
			return err
		}

		if account.IsNew {
			for _, exchange := range exchanges {
				balance := initialBalance[exchange]

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
					return err
				}
			}
		}

		return nil
	})

	return key.Token, err
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

	return uc.apikey.UpdateAccountKey(ctx, key)
}
