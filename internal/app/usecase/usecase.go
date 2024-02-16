package usecase

import "DemoExchange/internal/app/entities"

const lenBufferOrders = 100

type Config struct {
	KeyLimit    int
	MaxBalances map[entities.Coin]float64
}

type Usecase struct {
	cfg      Config
	account  AccountStorage
	apikey   APIKeyStorage
	wallet   WalletStorage
	order    OrderStorage
	position PositionStorage
	cache    Cache
	log      Logger
	chOrders chan interface{}
}

func New(cfg Config, account AccountStorage, apikey APIKeyStorage, wallet WalletStorage, order OrderStorage, position PositionStorage, cache Cache, log Logger) *Usecase {
	return &Usecase{
		cfg,
		account,
		apikey,
		wallet,
		order,
		position,
		cache,
		log,
		make(chan interface{}, lenBufferOrders),
	}
}
