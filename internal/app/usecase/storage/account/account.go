package account

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

func (s *Storage) InsertAccount(ctx context.Context, tx pgx.Tx, account *entities.Account) error {
	sql := `
		INSERT INTO account (account_uid, service, user_id, position_mode, position_type, create_ts, update_ts) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var err error

	if tx != nil {
		_, err = tx.Exec(ctx, sql, account.AccountUID, account.Service, account.UserID, account.PositionMode, account.PositionType, account.CreateTS, account.UpdateTS)
	} else {
		err = s.repo.Exec(ctx, sql, account.AccountUID, account.Service, account.UserID, account.PositionMode, account.PositionType, account.CreateTS, account.UpdateTS)
	}

	return err
}

func (s *Storage) UpdatePositionMode(ctx context.Context, account *entities.Account) error {
	sql := `UPDATE account SET position_mode = $2, update_ts = $3 WHERE account_uid = $1`

	return s.repo.Exec(ctx, sql, account.AccountUID, account.PositionMode, account.UpdateTS)
}

func (s *Storage) SelectAccount(ctx context.Context, tx pgx.Tx, service, userID string) (*entities.Account, error) {
	var account entities.Account

	sql := `
		SELECT account_uid, service, user_id, position_mode, position_type, create_ts, update_ts FROM account WHERE service = $1 AND user_id = $2
	`

	row := s.repo.QueryRow(ctx, sql, service, userID)

	err := row.Scan(&account.AccountUID, &account.Service, &account.UserID, &account.PositionMode, &account.PositionType, &account.CreateTS, &account.UpdateTS)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrAccountNotFound
		}
		return nil, err
	}

	return &account, nil
}

func (s *Storage) SelectAccountByUID(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID) (*entities.Account, error) {
	var account entities.Account

	sql := `
		SELECT account_uid, service, user_id, position_mode, position_type, create_ts, update_ts FROM account WHERE account_uid = $1
	`

	row := s.repo.QueryRow(ctx, sql, accountUID)

	err := row.Scan(&account.AccountUID, &account.Service, &account.UserID, &account.PositionMode, &account.PositionType, &account.CreateTS, &account.UpdateTS)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrAccountNotFound
		}
		return nil, err
	}

	return &account, nil
}
