package wallet

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"DemoExchange/internal/app/entities"
)

const codeDivisionByZero = "22012"

//lint:ignore ST1005 strings capitalized
var ErrInsufficientFunds = errors.New("Insufficient funds")

type Repository interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) error
}

type Storage struct {
	repo Repository
}

func New(repo Repository) *Storage {
	return &Storage{
		repo,
	}
}

func (s *Storage) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.repo.WithTx(ctx, fn)
}

func (s *Storage) SelectBalances(ctx context.Context, wallet entities.Wallet) (entities.Balances, error) {
	result := make(entities.Balances, 0)

	sql := `SELECT coin, total, hold FROM wallet WHERE exchange = $1 AND account_uid = $2`

	rows, err := s.repo.Query(ctx, sql, wallet.Exchange, wallet.AccountUID)
	if err != nil {
		return result, err
	}

	var (
		coin  entities.Coin
		total float64
		hold  float64
	)

	balances := make(entities.Balances)

	_, err = pgx.ForEachRow(rows, []any{&coin, &total, &hold}, func() error {
		balances[coin] = entities.Balance{
			Coin:  coin,
			Total: total,
			Hold:  hold,
		}

		return nil
	})

	return balances, err
}

func (s *Storage) AppendTotalCoin(ctx context.Context, wallet entities.Wallet) error {
	sql := `
		INSERT INTO wallet (exchange, account_uid, coin, total, update_ts) VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (exchange, account_uid, coin) DO UPDATE SET total = wallet.total + EXCLUDED.total
	`

	return s.repo.Exec(ctx, sql, wallet.Exchange, wallet.AccountUID, wallet.Balance.Coin, wallet.Balance.Total, wallet.UpdateTS)
}

func (s *Storage) SubtractTotalCoin(ctx context.Context, wallet entities.Wallet) error {
	sql := `
		UPDATE wallet SET total = total - $4, update_ts = $5
		WHERE exchange = $1 AND account_uid = $2 AND coin = $3
			AND	(1 = 1 / CASE WHEN total >= $4 THEN 1 ELSE 0 END)
			--AND	(1 = 1 / CASE WHEN total - hold >= $4 THEN 1 ELSE 0 END)
	`

	err := s.repo.Exec(ctx, sql, wallet.Exchange, wallet.AccountUID, wallet.Balance.Coin, wallet.Balance.Total, wallet.UpdateTS)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == codeDivisionByZero {
				return ErrInsufficientFunds
			}
		}
	}

	return err
}

func (s *Storage) SetHoldCoin(ctx context.Context, wallet entities.Wallet) error {
	sql := `
		UPDATE wallet SET hold = $4, update_ts = $5 WHERE exchange = $1 AND account_uid = $2 AND coin = $3
	`

	return s.repo.Exec(ctx, sql, wallet.Exchange, wallet.AccountUID, wallet.Balance.Coin, wallet.Balance.Hold, wallet.UpdateTS)
}
