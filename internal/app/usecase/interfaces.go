package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"

	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/usecase/trade"
)

type Tx interface {
	WithTX(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type AccountStorage interface {
	InsertAccount(ctx context.Context, tx pgx.Tx, account *entities.Account) error
	UpdatePositionMode(ctx context.Context, account *entities.Account) error
	SelectAccount(ctx context.Context, tx pgx.Tx, service, userID string) (*entities.Account, error)
	SelectAccountByUID(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID) (*entities.Account, error)
}

type APIKeyStorage interface {
	InsertAccountKey(ctx context.Context, tx pgx.Tx, key *entities.Key) error
	UpdateAccountKey(ctx context.Context, key *entities.Key) error
	SelectAccountKeys(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID) ([]entities.Key, error)
	SelectAccountUID(ctx context.Context, token entities.Token) (entities.AccountUID, error)
}

type WalletStorage interface {
	SelectBalances(ctx context.Context, tx pgx.Tx, wallet entities.Wallet) (entities.Balances, error)
	AppendTotalCoin(ctx context.Context, tx pgx.Tx, wallet entities.Wallet) error
	SubtractTotalCoin(ctx context.Context, tx pgx.Tx, wallet entities.Wallet) error
	SetHoldCoin(ctx context.Context, tx pgx.Tx, wallet entities.Wallet) error
}

type OrderStorage interface {
	InsertOrder(ctx context.Context, order *entities.Order) error
	SelectOrder(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, orderUID string) (*entities.Order, error)
	UpdateOrder(ctx context.Context, order *entities.Order) error
	SelectOrders(ctx context.Context, exchange entities.Exchange, accountUID entities.AccountUID, statuses []entities.OrderStatus, limit int) ([]*entities.Order, error)
	SelectAllPendingOrders(ctx context.Context) ([]*entities.Order, error)
}

type PositionStorage interface {
	InsertPosition(ctx context.Context, tx pgx.Tx, position *entities.Position) error
	UpdatePosition(ctx context.Context, tx pgx.Tx, position *entities.Position) error
	SelectPositionBySide(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID, symbol entities.Symbol, side entities.PositionSide) (*entities.Position, error)
	SelectPositionsBySymbol(ctx context.Context, tx pgx.Tx, accountUID entities.AccountUID, symbol entities.Symbol) (map[entities.PositionSide]*entities.Position, error)
	SelectAccountPositions(ctx context.Context, accountUID entities.AccountUID) ([]*entities.Position, error)
}

type Storage interface {
	APIKeyStorage
	WalletStorage
	OrderStorage
}

type Trade interface {
	Create(order *entities.Order) (trade.Trader, error)
	Set(order *entities.Order)
	Get(orderUID string) (*entities.Order, error)
	Delete(orderUID string)
	List() []*entities.Order
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}
