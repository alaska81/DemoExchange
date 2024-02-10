package usecase

import "DemoExchange/internal/app/entities"

const lenBufferOrders = 100

type Config struct {
	KeyLimit    int
	MaxBalances map[entities.Coin]float64
}

type Usecase struct {
	cfg       Config
	tx        Tx
	account   AccountStorage
	apikey    APIKeyStorage
	wallet    WalletStorage
	order     OrderStorage
	position  PositionStorage
	trade     Trade
	log       Logger
	chTraders chan interface{}
}

func New(cfg Config, tx Tx, account AccountStorage, apikey APIKeyStorage, wallet WalletStorage, order OrderStorage, position PositionStorage, trade Trade, log Logger) *Usecase {
	return &Usecase{
		cfg,
		tx,
		account,
		apikey,
		wallet,
		order,
		position,
		trade,
		log,
		make(chan interface{}, lenBufferOrders),
	}
}
