package transaction

import (
	"DemoExchange/internal/app/entities"
	"context"

	"github.com/jackc/pgx/v5"
)

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

func (s *Storage) InsertTransaction(ctx context.Context, transaction *entities.Transaction) error {
	sql := `
		INSERT INTO "transaction" (account_uid, transaction_uid, exchange, symbol, transaction_type, amount, create_ts) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	return s.repo.Exec(ctx, sql, transaction.AccountUID, transaction.TransactionID, transaction.Exchange, transaction.Symbol, transaction.TransactionType, transaction.Amount, transaction.CreateTS)
}

func (s *Storage) SelectAccountTransactions(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, filter entities.TransactionFilter) ([]*entities.Transaction, error) {
	sql := `
		SELECT account_uid, transaction_uid, exchange, symbol, transaction_type, amount, create_ts 
		FROM "transaction" 
		WHERE exchange = $1 AND account_uid = $2
			AND (transaction_type = $3 OR $3 = '')
			AND (create_ts >= $4 OR $4 = 0)
			AND (create_ts <= $5 OR $5 = 0)
		ORDER BY create_ts DESC
		LIMIT $6
	`
	var (
		rows pgx.Rows
		err  error
	)

	rows, err = s.repo.Query(ctx, sql, exchange, accountUID, filter.TransactionType, filter.From, filter.To, filter.Limit)
	if err != nil {
		return nil, err
	}

	var transaction entities.Transaction

	transactions := make([]*entities.Transaction, 0)

	_, err = pgx.ForEachRow(rows, []any{&transaction.AccountUID, &transaction.TransactionID, &transaction.Exchange, &transaction.Symbol, &transaction.TransactionType, &transaction.Amount, &transaction.CreateTS}, func() error {
		transaction := transaction
		transactions = append(transactions, &transaction)
		return nil
	})

	return transactions, err
}
