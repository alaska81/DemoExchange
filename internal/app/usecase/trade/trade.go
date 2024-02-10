package trade

import (
	"context"
	"errors"
	"time"

	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/tickers"
)

var processTimeout = 5 * time.Second

//lint:file-ignore ST1005 strings capitalized

var (
	ErrOrderTypeNotValid   = errors.New("Order Type not valid")
	ErrOrderNotFound       = errors.New("Order not found")
	ErrExchangeIsNotValid  = errors.New("Exchange is not valid")
	ErrSymbolIsNotValid    = errors.New("Symbol is not valid")
	ErrOrderSideIsNotValid = errors.New("Order Side is not valid")
	ErrAmountIsNotValid    = errors.New("Amount is not valid")
	ErrPriceIsNotValid     = errors.New("Price is not valid")
	ErrMarketPriceIsWrong  = errors.New("Market price is wrong")
)

type Trader interface {
	Process(ctx context.Context) <-chan *entities.Order
}

type Storage interface {
	Set(order *entities.Order)
	Get(orderUID string) (*entities.Order, bool)
	Delete(orderUID string)
	List() []*entities.Order
}

type trade struct {
	storage Storage
}

func New(storage Storage) *trade {
	return &trade{
		storage: storage,
	}
}

func (t *trade) Create(order *entities.Order) (Trader, error) {
	var trader Trader

	switch order.Type {
	case entities.OrderTypeMarket:
		marketOrder := NewMarketOrder(order)

		err := marketOrder.Validate()
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ticker, err := tickers.New().GetTickerWithContext(ctx, order.Exchange.Name(), order.Symbol.String())
		if err != nil {
			return nil, ErrMarketPriceIsWrong
		}

		marketOrder.SetPrice(ticker.Last)
		trader = marketOrder

	case entities.OrderTypeLimit:
		limitOrder := NewLimitOrder(order)

		err := limitOrder.Validate()
		if err != nil {
			return nil, err
		}

		trader = limitOrder

	default:
		return nil, ErrOrderTypeNotValid
	}

	return trader, nil
}

func (t *trade) Set(order *entities.Order) {
	t.storage.Set(order)
}

func (t *trade) Get(orderUID string) (*entities.Order, error) {
	order, ok := t.storage.Get(orderUID)
	if !ok {
		return nil, ErrOrderNotFound
	}

	return order, nil
}

func (t *trade) Delete(orderUID string) {
	t.storage.Delete(orderUID)
}

func (t *trade) List() []*entities.Order {
	return t.storage.List()
}

func SetProcessTimeout(timeout time.Duration) {
	processTimeout = timeout
}
