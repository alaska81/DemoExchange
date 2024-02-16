package position

import (
	"context"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/apperror"
	"DemoExchange/internal/app/entities"
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

func (s *Storage) InsertPosition(ctx context.Context, position *entities.Position) error {
	sql := `
		INSERT INTO "position" (account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, margin, hold_amount, create_ts, update_ts) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING amount, price, hold_amount
	`

	row := s.repo.QueryRow(ctx, sql, position.AccountUID, position.PositionUID, position.Exchange, position.Symbol, position.Mode, position.MarginType, position.Leverage, position.Side, position.Amount, position.Price, position.Margin, position.HoldAmount, position.CreateTS, position.UpdateTS)

	return row.Scan(&position.Amount, &position.Price, &position.HoldAmount)
}

func (s *Storage) UpdatePosition(ctx context.Context, position *entities.Position) error {
	sql := `
		UPDATE "position" SET amount = $2, price = $3, margin = $4, hold_amount = $5, position_type = $6, leverage = $7, update_ts = $8 WHERE position_uid = $1
	`

	return s.repo.Exec(ctx, sql, position.PositionUID, position.Amount, position.Price, position.Margin, position.HoldAmount, position.MarginType, position.Leverage, position.UpdateTS)
}

func (s *Storage) SelectPositionBySide(ctx context.Context, accountUID entities.AccountUID, symbol entities.Symbol, side entities.PositionSide) (*entities.Position, error) {
	var position entities.Position

	sql := `
		SELECT account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, margin, hold_amount, create_ts, update_ts 
		FROM "position" 
		WHERE account_uid = $1 AND symbol = $2 AND side = $3
	`

	row := s.repo.QueryRow(ctx, sql, accountUID, symbol, side)

	err := row.Scan(&position.AccountUID, &position.PositionUID, &position.Exchange, &position.Symbol, &position.Mode, &position.MarginType, &position.Leverage, &position.Side, &position.Amount, &position.Price, &position.Margin, &position.HoldAmount, &position.CreateTS, &position.UpdateTS)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrPositionNotFound
		}
		return nil, apperror.ErrRequestError
	}

	return &position, nil
}

func (s *Storage) SelectPositionsBySymbol(ctx context.Context, accountUID entities.AccountUID, symbol entities.Symbol) (map[entities.PositionSide]*entities.Position, error) {
	sql := `
		SELECT account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, margin, hold_amount, create_ts, update_ts 
		FROM "position" 
		WHERE account_uid = $1 AND symbol = $2
	`

	rows, err := s.repo.Query(ctx, sql, accountUID, symbol)
	if err != nil {
		return nil, err
	}

	var position entities.Position

	positions := make(map[entities.PositionSide]*entities.Position)

	_, err = pgx.ForEachRow(rows, []any{&position.AccountUID, &position.PositionUID, &position.Exchange, &position.Symbol, &position.Mode, &position.MarginType, &position.Leverage, &position.Side, &position.Amount, &position.Price, &position.Margin, &position.HoldAmount, &position.CreateTS, &position.UpdateTS}, func() error {
		position := position
		positions[position.Side] = &position
		return nil
	})

	return positions, err
}

func (s *Storage) SelectAccountPositions(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) ([]*entities.Position, error) {
	sql := `
		SELECT account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, margin, hold_amount, create_ts, update_ts 
		FROM "position" 
		WHERE exchange = $1 AND account_uid = $2
		ORDER BY create_ts DESC
	`
	var (
		rows pgx.Rows
		err  error
	)

	rows, err = s.repo.Query(ctx, sql, exchange, accountUID)
	if err != nil {
		return nil, err
	}

	var position entities.Position

	positions := make([]*entities.Position, 0)

	_, err = pgx.ForEachRow(rows, []any{&position.AccountUID, &position.PositionUID, &position.Exchange, &position.Symbol, &position.Mode, &position.MarginType, &position.Leverage, &position.Side, &position.Amount, &position.Price, &position.Margin, &position.HoldAmount, &position.CreateTS, &position.UpdateTS}, func() error {
		position := position
		positions = append(positions, &position)
		return nil
	})

	return positions, err
}

func (s *Storage) SelectAccountOpenPositions(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) ([]*entities.Position, error) {
	sql := `
		SELECT account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, margin, hold_amount, create_ts, update_ts 
		FROM "position" 
		WHERE exchange = $1 AND account_uid = $2 AND (hold_amount > 0 OR amount > 0)
	`
	var (
		rows pgx.Rows
		err  error
	)

	rows, err = s.repo.Query(ctx, sql, exchange, accountUID)
	if err != nil {
		return nil, err
	}

	var position entities.Position

	positions := make([]*entities.Position, 0)

	_, err = pgx.ForEachRow(rows, []any{&position.AccountUID, &position.PositionUID, &position.Exchange, &position.Symbol, &position.Mode, &position.MarginType, &position.Leverage, &position.Side, &position.Amount, &position.Price, &position.Margin, &position.HoldAmount, &position.CreateTS, &position.UpdateTS}, func() error {
		position := position
		positions = append(positions, &position)
		return nil
	})

	return positions, err
}
