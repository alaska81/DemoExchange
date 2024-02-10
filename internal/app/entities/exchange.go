package entities

type Exchange string

const (
	ExchangeSpot    Exchange = "demo_spot"
	ExchangeFutures Exchange = "demo_futures"
)

var exchanges = map[Exchange]string{
	ExchangeSpot:    "binance",
	ExchangeFutures: "binance_futures",
}

func (e Exchange) Name() string {
	return exchanges[e]
}
