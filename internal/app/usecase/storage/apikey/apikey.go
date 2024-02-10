package apikey

import (
	"context"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
)

type Repository interface {
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

func (s *Storage) InsertAccountKey(ctx context.Context, tx pgx.Tx, key *entities.Key) error {
	sql := `
		INSERT INTO apikey (token, account_uid, create_ts, update_ts) VALUES ($1, $2, $3, $4)
	`

	var err error

	if tx != nil {
		_, err = tx.Exec(ctx, sql, key.Token, key.AccountUID, key.CreateTS, key.UpdateTS)
	} else {
		err = s.repo.Exec(ctx, sql, key.Token, key.AccountUID, key.CreateTS, key.UpdateTS)
	}

	return err
}

func (s *Storage) UpdateAccountKey(ctx context.Context, key *entities.Key) error {
	sql := `UPDATE apikey SET disabled = $2, update_ts = $3 WHERE token = $1`

	return s.repo.Exec(ctx, sql, key.Token, key.Disabled, key.UpdateTS)
}

func (s *Storage) SelectAccountKeys(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID) ([]entities.Key, error) {
	result := make([]entities.Key, 0)

	sql := `SELECT token, create_ts, update_ts FROM apikey WHERE account_uid = $1 AND disabled = false`

	var (
		rows pgx.Rows
		err  error
	)

	if tx != nil {
		rows, err = tx.Query(ctx, sql, accountUID)
	} else {
		rows, err = s.repo.Query(ctx, sql, accountUID)
	}

	if err != nil {
		return result, err
	}

	var (
		token    entities.Token
		createTS int64
		updateTS int64
	)

	_, err = pgx.ForEachRow(rows, []any{&token, &createTS, &updateTS}, func() error {
		result = append(result, entities.Key{
			Token:      token,
			AccountUID: accountUID,
			CreateTS:   createTS,
			UpdateTS:   updateTS,
		})
		return nil
	})

	return result, err
}

func (s *Storage) SelectAccountUID(ctx context.Context, token entities.Token) (entities.AccountUID, error) {
	sql := `SELECT account_uid FROM apikey WHERE token = $1`

	row := s.repo.QueryRow(ctx, sql, token)

	var accountUID entities.AccountUID

	err := row.Scan(&accountUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", apperror.ErrTokenNotFound
		}
		return "", apperror.ErrRequestError
	}

	return accountUID, nil
}
