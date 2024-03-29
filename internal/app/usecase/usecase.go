package usecase

import (
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/markets"
	"DemoExchange/internal/app/tickers"
	"DemoExchange/internal/app/usecase/cache"
	"DemoExchange/internal/app/usecase/orders"
	"DemoExchange/internal/app/usecase/repo/account"
	"DemoExchange/internal/app/usecase/repo/apikey"
	"DemoExchange/internal/app/usecase/repo/order"
	"DemoExchange/internal/app/usecase/repo/position"
	"DemoExchange/internal/app/usecase/repo/transaction"
	"DemoExchange/internal/app/usecase/repo/wallet"
	"context"
)

const lenBufferOrders = 100

var exchanges = []entities.Exchange{entities.ExchangeSpot, entities.ExchangeFutures}

type Config struct {
	KeyLimit    int
	MaxBalances map[entities.Coin]float64
}

type Usecase struct {
	cfg Config

	account     AccountStorage
	apikey      APIKeyStorage
	wallet      WalletStorage
	order       OrderStorage
	position    PositionStorage
	transaction TransactionStorage

	cacheOrders    Cache[string, *entities.Order]
	cachePositions Cache[string, *entities.Position]

	chOrders    chan *orders.Order
	chPositions chan *entities.Position

	tickers Tickers
	markets Markets
	log     Logger
}

func (uc *Usecase) GetMarketWithContext(ctx context.Context, exchange string, market string) (markets.Market, error) {
	return uc.markets.GetMarketWithContext(ctx, exchange, market)
}

func (uc *Usecase) GetTickerWithContext(ctx context.Context, exchange string, market string) (tickers.Ticker, error) {
	return uc.tickers.GetTickerWithContext(ctx, exchange, market)
}

// New creates new usecase
func New(cfg Config, repo Connection, tickers Tickers, markets Markets, log Logger) *Usecase {
	return &Usecase{
		cfg: cfg,

		account:     account.New(repo),
		apikey:      apikey.New(repo),
		wallet:      wallet.New(repo),
		order:       order.New(repo),
		position:    position.New(repo),
		transaction: transaction.New(repo),

		cacheOrders:    cache.New[string, *entities.Order](log),
		cachePositions: cache.New[string, *entities.Position](log),

		chOrders:    make(chan *orders.Order, lenBufferOrders),
		chPositions: make(chan *entities.Position),

		tickers: tickers,
		markets: markets,
		log:     log,
	}
}
