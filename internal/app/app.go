package app

import (
	"context"
	"fmt"
	"time"

	"DemoExchange/internal/adapters/apiservice"
	"DemoExchange/internal/adapters/postgres"
	"DemoExchange/internal/adapters/webserver"
	"DemoExchange/internal/app/entities"
	"DemoExchange/internal/app/markets"
	"DemoExchange/internal/app/orderbook"
	"DemoExchange/internal/app/tickers"
	"DemoExchange/internal/app/usecase"
	"DemoExchange/internal/config"
	"DemoExchange/internal/logger"
	"DemoExchange/migrator"
)

type App struct {
	cfg       *config.Config
	pool      *postgres.Connection
	markets   *markets.Service
	tickers   *tickers.Service
	orderbook *orderbook.Service
	usecase   *usecase.Usecase
	webserver *webserver.Server
	log       *logger.Logger
}

func New() (*App, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	log, err := logger.GetInstance(
		logger.Config{
			Level: cfg.Logger.Level,
			Path:  cfg.Logger.Path,
			File:  cfg.Logger.File,
		},
	)
	if err != nil {
		return nil, err
	}

	repo, err := postgres.NewConnection(
		postgres.Config{
			Host:         cfg.DB.Host,
			Port:         cfg.DB.Port,
			User:         cfg.DB.User,
			Password:     cfg.DB.Password,
			Database:     cfg.DB.Database,
			MinOpenConns: cfg.DB.MinOpenConns,
			MaxOpenConns: cfg.DB.MaxOpenConns,
		},
	)
	if err != nil {
		return nil, err
	}

	apiclient := apiservice.NewClient(
		apiservice.Config{
			Address: cfg.APIService.Address,
			Timeout: time.Duration(cfg.APIService.Timeout) * time.Second,
		},
	)
	marketsService := markets.NewService(apiclient, log)
	tickersService := tickers.NewService(apiclient, log)
	orderbookService := orderbook.NewService(apiclient, log)

	cfgUsecase := usecase.Config{
		KeyLimit: cfg.Service.KeyLimit,
		MaxBalances: map[entities.Coin]float64{
			"USDT": 5000,
		},
	}

	tickers := tickers.New()
	markets := markets.New()

	usecase := usecase.New(cfgUsecase, repo, tickers, markets, log)

	webserver, err := webserver.New(
		webserver.Config{
			Host:               cfg.WebServer.Host,
			Port:               cfg.WebServer.Port,
			PortTLS:            cfg.WebServer.PortTLS,
			Timeout:            time.Duration(cfg.WebServer.Timeout) * time.Second,
			CertCrt:            cfg.WebServer.CertCrt,
			CertKey:            cfg.WebServer.CertKey,
			AllowServiceTokens: cfg.WebServer.AllowServiceTokens,
		},
		markets,
		tickers,
		orderbookService,
		usecase,
		log,
	)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:       cfg,
		pool:      repo,
		usecase:   usecase,
		webserver: webserver,
		markets:   marketsService,
		tickers:   tickersService,
		orderbook: orderbookService,
		log:       log,
	}, nil
}

func (a *App) Run(ctx context.Context, cancel context.CancelFunc) error {
	err := a.pool.NewPool(ctx)
	if err != nil {
		return fmt.Errorf("new pool: %w", err)
	}

	err = migrator.Migrate(a.pool.GetPool())
	if err != nil {
		return fmt.Errorf("migrator: %w", err)
	}

	<-a.markets.Process(ctx)
	<-a.tickers.Process(ctx)

	go a.usecase.ProcessOrders(ctx)
	go a.usecase.ProcessPositions(ctx)

	go func() {
		err := a.webserver.Start(ctx)
		if err != nil {
			cancel()
		}
	}()

	err = a.usecase.ProcessPendingOrders(ctx)
	if err != nil {
		return fmt.Errorf("process pending orders: %w", err)
	}

	err = a.usecase.ProcessOpenPositions(ctx)
	if err != nil {
		return fmt.Errorf("process open positions: %w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.pool.Close()
}
