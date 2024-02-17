package usecase

import (
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/tickers"
	"DemoExchange/internal/app/usecase/cache"
	"DemoExchange/internal/app/usecase/orders"
	"DemoExchange/internal/app/usecase/repo/account"
	"DemoExchange/internal/app/usecase/repo/apikey"
	"DemoExchange/internal/app/usecase/repo/order"
	"DemoExchange/internal/app/usecase/repo/position"
	"DemoExchange/internal/app/usecase/repo/wallet"
)

const lenBufferOrders = 100

type Config struct {
	KeyLimit    int
	MaxBalances map[entities.Coin]float64
}

type Usecase struct {
	cfg            Config
	account        AccountStorage
	apikey         APIKeyStorage
	wallet         WalletStorage
	order          OrderStorage
	position       PositionStorage
	cacheOrders    Cache
	cachePositions Cache
	log            Logger
	chOrders       chan *orders.Order
	chPositions    chan *entities.Position
	tickers        *tickers.Receiver
}

func New(cfg Config, repo Connection, log Logger) *Usecase {
	account := account.New(repo)
	apikey := apikey.New(repo)
	wallet := wallet.New(repo)
	order := order.New(repo)
	position := position.New(repo)

	return &Usecase{
		cfg,
		account,
		apikey,
		wallet,
		order,
		position,
		cache.New(),
		cache.New(),
		log,
		make(chan *orders.Order, lenBufferOrders),
		make(chan *entities.Position),
		tickers.New(),
	}
}
