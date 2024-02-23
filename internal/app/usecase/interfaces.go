package usecase

import (
	"context"

	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/usecase/orders"

	"github.com/jackc/pgx/v5"
)

type Connection interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	Exec(ctx context.Context, sql string, args ...any) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type AccountStorage interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	InsertAccount(ctx context.Context, account *entities.Account) error
	UpdatePositionMode(ctx context.Context, account *entities.Account) error
	SelectAccount(ctx context.Context, service, userID string) (*entities.Account, error)
	SelectAccountByUID(ctx context.Context, accountUID entities.AccountUID) (*entities.Account, error)
}

type APIKeyStorage interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	InsertAccountKey(ctx context.Context, key *entities.Key) error
	UpdateAccountKey(ctx context.Context, key *entities.Key) error
	SelectAccountKeys(ctx context.Context, accountUID entities.AccountUID) ([]entities.Key, error)
	SelectAccountUID(ctx context.Context, token entities.Token) (entities.AccountUID, error)
}

type WalletStorage interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	SelectBalances(ctx context.Context, wallet entities.Wallet) (entities.Balances, error)
	AppendTotalCoin(ctx context.Context, wallet entities.Wallet) error
	SubtractTotalCoin(ctx context.Context, wallet entities.Wallet) error
	SetHoldCoin(ctx context.Context, wallet entities.Wallet) error
}

type OrderStorage interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	InsertOrder(ctx context.Context, order *entities.Order) error
	SelectOrder(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, orderUID string) (*entities.Order, error)
	UpdateOrder(ctx context.Context, order *entities.Order) error
	SelectOrders(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, statuses []entities.OrderStatus, limit int) ([]*entities.Order, error)
	SelectPendingOrders(ctx context.Context) ([]*entities.Order, error)
	SelectPendingOrdersBySymbol(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, symbol *entities.Symbol) ([]*entities.Order, error)
}

type PositionStorage interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	InsertPosition(ctx context.Context, position *entities.Position) error
	UpdatePosition(ctx context.Context, position *entities.Position) error
	SelectPositionBySide(ctx context.Context, accountUID entities.AccountUID, symbol entities.Symbol, side entities.PositionSide) (*entities.Position, error)
	SelectPositionsBySymbol(ctx context.Context, accountUID entities.AccountUID, symbol entities.Symbol) (map[entities.PositionSide]*entities.Position, error)
	SelectAccountPositions(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) ([]*entities.Position, error)
	SelectAccountOpenPositions(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID) ([]*entities.Position, error)
	SelectOpenPositions(ctx context.Context) ([]*entities.Position, error)
}

type TransactionStorage interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	InsertTransaction(ctx context.Context, transaction *entities.Transaction) error
	SelectAccountTransactions(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, filter entities.TransactionFilter) ([]*entities.Transaction, error)
}

type Cache interface {
	Set(uid string, value any)
	Get(uid string) (any, bool)
	Delete(uid string)
	List() []any
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type Order interface {
	GetOrder() *entities.Order
	Validate() error
	Process(ctx context.Context) (<-chan entities.OrderStatus, error)
	HoldBalance(ctx context.Context, uc orders.Usecase, log orders.Logger) error
	UnholdBalance(ctx context.Context, uc orders.Usecase, log orders.Logger) error
	AppendBalance(ctx context.Context, uc orders.Usecase, log orders.Logger) error
}
