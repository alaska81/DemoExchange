package tx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Storage struct {
	repo Repository
}

func New(repo Repository) *Storage {
	return &Storage{
		repo,
	}
}

func (s *Storage) WithTX(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := s.repo.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
