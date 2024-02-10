package tickers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"DemoExchange/internal/adapters/apiservice"
)

var (
	tickersMap sync.Map
	exchanges  = []string{"binance", "binance_futures"}
	timeout    = 3 * time.Second
)

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

func (s *Service) SetExchanges(e []string) {
	exchanges = e
}

func (s *Service) SetTimeout(t time.Duration) {
	timeout = t
}

func (s *Service) Process(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(exchanges))

		for _, exchange := range exchanges {
			go func(exchange string) {
				s.log.Info(fmt.Sprintf("tickers:Process Start %s", exchange))

				once := sync.Once{}

				req := s.api.GetRequest("GET", fmt.Sprintf("/public/tickers/%s", exchange), nil)

				for {
					var tickers Tickers
					if err := req.Do(ctx, &tickers); err != nil {
						if errors.Is(err, context.Canceled) {
							once.Do(func() {
								wg.Done()
							})
							s.log.Info(fmt.Sprintf("tickers:Process Stop %s", exchange))
							return
						}

						time.Sleep(timeout)
						continue
					}

					tickersMap.Store(exchange, tickers)

					once.Do(func() {
						wg.Done()
					})

					time.Sleep(timeout)
				}
			}(exchange)
		}

		wg.Wait()

		s.log.Info("tickers:Process Begin")

		close(ch)
	}()

	return ch
}
