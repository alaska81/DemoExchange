package orderbook

import (
	"context"
	"errors"
	"fmt"

	"DemoExchange/internal/adapters/apiservice"
)

// var (
// 	orderbookMap sync.Map
// 	exchanges    = []string{"binance", "binance_futures"}
// 	timeout      = 3 * time.Second
// )

var ErrOrderbookIsEmpty = errors.New("Orderbook is empty")

type Client interface {
	GetRequest(method string, query string, payload any) apiservice.Request
}

type Logger interface {
	Info(args ...interface{})
}

type Service struct {
	api Client
	log Logger
}

func NewService(api Client, log Logger) *Service {
	return &Service{
		api,
		log,
	}
}

// func (s *Service) SetExchanges(e []string) {
// 	exchanges = e
// }

// func (s *Service) SetTimeout(t time.Duration) {
// 	timeout = t
// }

func (s *Service) GetOrderbook(ctx context.Context, exchange, symbol, limit string) (*Orderbook, error) {
	return s.request(ctx, exchange, symbol, limit)
}

func (s *Service) request(ctx context.Context, exchange, symbol, limit string) (*Orderbook, error) {
	var orderbook Orderbook

	req := s.api.GetRequest("GET", fmt.Sprintf("/public/orderbook/%s/%s/%s", exchange, symbol, limit), nil)
	if err := req.Do(ctx, &orderbook); err != nil {
		if errors.Is(err, context.Canceled) {
			err = ErrOrderbookIsEmpty
		}
		return nil, err
	}

	return &orderbook, nil
}

// func (s *Service) Process(ctx context.Context) <-chan struct{} {
// 	ch := make(chan struct{})

// 	go func() {
// 		wg := sync.WaitGroup{}
// 		wg.Add(len(exchanges))

// 		for _, exchange := range exchanges {
// 			go func(exchange string) {
// 				s.log.Info(fmt.Sprintf("orderbook:Process Start %s", exchange))

// 				once := sync.Once{}

// 				req := s.api.GetRequest("GET", fmt.Sprintf("/public/orderbook/%s", exchange), nil)

// 				for {
// 					var orderbook Orderbook
// 					if err := req.Do(ctx, &orderbook); err != nil {
// 						if errors.Is(err, context.Canceled) {
// 							once.Do(func() {
// 								wg.Done()
// 							})
// 							s.log.Info(fmt.Sprintf("orderbook:Process Stop %s", exchange))
// 							return
// 						}

// 						time.Sleep(timeout)
// 						continue
// 					}

// 					orderbookMap.Store(exchange, orderbook)

// 					once.Do(func() {
// 						wg.Done()
// 					})

// 					time.Sleep(timeout)
// 				}
// 			}(exchange)
// 		}

// 		wg.Wait()

// 		s.log.Info("orderbook:Process Begin")

// 		close(ch)
// 	}()

// 	return ch
// }
