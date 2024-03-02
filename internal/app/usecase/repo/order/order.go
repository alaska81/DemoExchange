package order

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/entities"
)

//lint:ignore ST1005 strings capitalized
var ErrOrderNotFound = errors.New("Order not found")

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

func (s *Storage) InsertOrder(ctx context.Context, order *entities.Order) error {
	sql := `
		INSERT INTO "order" (account_uid, order_uid, exchange, symbol, type, position_side, side, amount, price, fee, fee_coin, reduce_only, status, create_ts, update_ts) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING amount, price
	`
	row := s.repo.QueryRow(ctx, sql, order.AccountUID, order.OrderUID, order.Exchange, order.Symbol, order.Type, order.PositionSide, order.Side, order.Amount, order.Price, order.Fee, order.FeeCoin, order.ReduceOnly, order.Status, order.CreateTS, order.UpdateTS)

	return row.Scan(&order.Amount, &order.Price)
}

func (s *Storage) SelectOrder(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, orderUID string) (*entities.Order, error) {
	var order entities.Order

	sql := `
		SELECT account_uid, order_uid, exchange, symbol, type, position_side, side, amount, price, fee, fee_coin, reduce_only, status, create_ts, update_ts 
		FROM "order" 
		WHERE exchange = $1 AND account_uid = $2 AND order_uid = $3
	`

	row := s.repo.QueryRow(ctx, sql, exchange, accountUID, orderUID)

	err := row.Scan(&order.AccountUID, &order.OrderUID, &order.Exchange, &order.Symbol, &order.Type, &order.PositionSide, &order.Side, &order.Amount, &order.Price, &order.Fee, &order.FeeCoin, &order.ReduceOnly, &order.Status, &order.CreateTS, &order.UpdateTS)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (s *Storage) UpdateOrder(ctx context.Context, order *entities.Order) error {
	sql := `
		UPDATE "order" SET status = $3, error = $4, price = $5, fee = $6, fee_coin = $7, update_ts = $8 WHERE account_uid = $1 AND order_uid = $2
	`
	return s.repo.Exec(ctx, sql, order.AccountUID, order.OrderUID, order.Status, order.Error, order.Price, order.Fee, order.FeeCoin, order.UpdateTS)
}

func (s *Storage) SelectOrders(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, statuses []entities.OrderStatus, limit int) ([]*entities.Order, error) {
	sql := `
		SELECT account_uid, order_uid, exchange, symbol, type, position_side, side, amount, price, fee, fee_coin, reduce_only, status, create_ts, update_ts 
		FROM "order" 
		WHERE exchange = $1 AND account_uid = $2 AND (status = ANY(string_to_array($3, ',')::text[]) OR $3 = '')
		ORDER BY create_ts DESC
		LIMIT $4
	`
	var (
		rows pgx.Rows
		err  error
	)

	statusString := entities.StatusArrayToString(statuses)
	rows, err = s.repo.Query(ctx, sql, exchange, accountUID, statusString, limit)
	if err != nil {
		return nil, err
	}

	var order entities.Order

	orders := make([]*entities.Order, 0)

	_, err = pgx.ForEachRow(rows, []any{&order.AccountUID, &order.OrderUID, &order.Exchange, &order.Symbol, &order.Type, &order.PositionSide, &order.Side, &order.Amount, &order.Price, &order.Fee, &order.FeeCoin, &order.ReduceOnly, &order.Status, &order.CreateTS, &order.UpdateTS}, func() error {
		order := order
		orders = append(orders, &order)
		return nil
	})

	return orders, err
}

func (s *Storage) SelectPendingOrders(ctx context.Context) ([]*entities.Order, error) {
	sql := `
		SELECT account_uid, order_uid, exchange, symbol, type, position_side, side, amount, price, fee, fee_coin, reduce_only, status, create_ts, update_ts 
		FROM "order" 
		WHERE status IN ($1, $2)
	`
	var (
		rows pgx.Rows
		err  error
	)

	rows, err = s.repo.Query(ctx, sql, entities.OrderStatusNew, entities.OrderStatusPending)
	if err != nil {
		return nil, err
	}

	var (
		order  entities.Order
		orders []*entities.Order
	)

	_, err = pgx.ForEachRow(rows, []any{&order.AccountUID, &order.OrderUID, &order.Exchange, &order.Symbol, &order.Type, &order.PositionSide, &order.Side, &order.Amount, &order.Price, &order.Fee, &order.FeeCoin, &order.ReduceOnly, &order.Status, &order.CreateTS, &order.UpdateTS}, func() error {
		order := order
		orders = append(orders, &order)
		return nil
	})

	return orders, err
}

func (s *Storage) SelectPendingOrdersBySymbol(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol *entities.Symbol) ([]*entities.Order, error) {
	sql := `
		SELECT account_uid, order_uid, exchange, symbol, type, position_side, side, amount, price, fee, fee_coin, reduce_only, status, create_ts, update_ts 
		FROM "order" 
		WHERE exchange = $1 AND account_uid = $2 
			AND (symbol = $3 OR $3 IS NULL)
			AND status = $4
	`
	var (
		rows pgx.Rows
		err  error
	)

	rows, err = s.repo.Query(ctx, sql, exchange, accountUID, symbol, entities.OrderStatusPending)
	if err != nil {
		return nil, err
	}

	var order entities.Order

	orders := make([]*entities.Order, 0)

	_, err = pgx.ForEachRow(rows, []any{&order.AccountUID, &order.OrderUID, &order.Exchange, &order.Symbol, &order.Type, &order.PositionSide, &order.Side, &order.Amount, &order.Price, &order.Fee, &order.FeeCoin, &order.ReduceOnly, &order.Status, &order.CreateTS, &order.UpdateTS}, func() error {
		order := order
		orders = append(orders, &order)
		return nil
	})

	return orders, err
}
