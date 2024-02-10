package position

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

func (s *Storage) InsertPosition(ctx context.Context, tx pgx.Tx, position *entities.Position) error {
	sql := `
		INSERT INTO "position" (account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, hold_amount, status, create_ts, update_ts) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING amount, price, hold_amount
	`

	var row pgx.Row

	if tx != nil {
		row = tx.QueryRow(ctx, sql, position.AccountUID, position.PositionUID, position.Exchange, position.Symbol, position.Mode, position.Type, position.Leverage, position.Side, position.Amount, position.Price, position.HoldAmount, position.Status, position.CreateTS, position.UpdateTS)
	} else {
		row = s.repo.QueryRow(ctx, sql, position.AccountUID, position.PositionUID, position.Exchange, position.Symbol, position.Mode, position.Type, position.Leverage, position.Side, position.Amount, position.Price, position.HoldAmount, position.Status, position.CreateTS, position.UpdateTS)
	}

	return row.Scan(&position.Amount, &position.Price, &position.HoldAmount)
}

func (s *Storage) SelectPositionByUID(ctx context.Context, accountUID entities.AccountUID, positionUID string) (*entities.Position, error) {
	var position entities.Position

	sql := `
		SELECT account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, hold_amount, status, create_ts, update_ts 
		FROM "position" 
		WHERE account_uid = $1 AND position_uid = $2
	`

	row := s.repo.QueryRow(ctx, sql, accountUID, positionUID)

	err := row.Scan(&position.AccountUID, &position.PositionUID, &position.Exchange, &position.Symbol, &position.Mode, &position.Type, &position.Leverage, &position.Side, &position.Amount, &position.Price, &position.HoldAmount, &position.Status, &position.CreateTS, &position.UpdateTS)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrPositionNotFound
		}
		return nil, apperror.ErrRequestError
	}

	return &position, nil
}

func (s *Storage) SelectOpenPositionBySymbolSide(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID, symbol entities.Symbol, side entities.PositionSide) (*entities.Position, error) {
	var position entities.Position

	sql := `
		SELECT account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, hold_amount, status, create_ts, update_ts 
		FROM "position" 
		WHERE account_uid = $1 AND symbol = $2 AND side = $3 AND status = $4
	`

	var row pgx.Row

	if tx != nil {
		row = tx.QueryRow(ctx, sql, accountUID, symbol, side, entities.PositionStatusOpen)
	} else {
		row = s.repo.QueryRow(ctx, sql, accountUID, symbol, side, entities.PositionStatusOpen)
	}

	err := row.Scan(&position.AccountUID, &position.PositionUID, &position.Exchange, &position.Symbol, &position.Mode, &position.Type, &position.Leverage, &position.Side, &position.Amount, &position.Price, &position.HoldAmount, &position.Status, &position.CreateTS, &position.UpdateTS)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.ErrPositionNotFound
		}
		return nil, apperror.ErrRequestError
	}

	return &position, nil
}

func (s *Storage) UpdatePosition(ctx context.Context, tx pgx.Tx, position *entities.Position) error {
	sql := `
		UPDATE "position" SET status = $3, amount = $4, price = $5, hold_amount = $6, update_ts = $7 WHERE account_uid = $1 AND position_uid = $2
	`

	var err error

	if tx != nil {
		_, err = tx.Exec(ctx, sql, position.AccountUID, position.PositionUID, position.Status, position.Amount, position.Price, position.HoldAmount, position.UpdateTS)
	} else {
		err = s.repo.Exec(ctx, sql, position.AccountUID, position.PositionUID, position.Status, position.Amount, position.Price, position.HoldAmount, position.UpdateTS)
	}

	return err
}

func (s *Storage) SelectOpenPositions(ctx context.Context, accountUID entities.AccountUID) ([]*entities.Position, error) {
	sql := `
		SELECT account_uid, position_uid, exchange, symbol, position_mode, position_type, leverage, side, amount, price, status, create_ts, update_ts 
		FROM "position" 
		WHERE account_uid = $1 AND status = $2
		ORDER BY create_ts DESC
	`
	var (
		rows pgx.Rows
		err  error
	)

	rows, err = s.repo.Query(ctx, sql, accountUID, entities.PositionStatusOpen)
	if err != nil {
		return nil, err
	}

	var position entities.Position

	positions := make([]*entities.Position, 0)

	_, err = pgx.ForEachRow(rows, []any{&position.AccountUID, &position.PositionUID, &position.Exchange, &position.Symbol, &position.Mode, &position.Type, &position.Leverage, &position.Side, &position.Amount, &position.Price, &position.Status, &position.CreateTS, &position.UpdateTS}, func() error {
		position := position
		positions = append(positions, &position)
		return nil
	})

	return positions, err
}
