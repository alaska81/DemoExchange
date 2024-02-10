package webserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	Host               string
	Port               string
	PortTLS            string
	Timeout            time.Duration
	CertCrt            string
	CertKey            string
	AllowServiceTokens []string
}

type Server struct {
	cfg       Config
	markets   Markets
	tickers   Tickers
	orderbook Orderbook
	usecase   Usecase
	log       Logger

	srv    *http.Server
	srvTLS *http.Server
}

func New(cfg Config, markets Markets, tickers Tickers, orderbook Orderbook, usecase Usecase, log Logger) (*Server, error) {
	s := Server{
		cfg:       cfg,
		markets:   markets,
		tickers:   tickers,
		orderbook: orderbook,
		usecase:   usecase,
		log:       log,
	}

	gin.SetMode(gin.ReleaseMode)

	handler := s.NewRoutes().Handler()

	s.srv = &http.Server{
		Addr:        fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:     handler,
		ReadTimeout: cfg.Timeout,
		// ReadHeaderTimeout: time.Duration(Conf.TimeoutGin) * time.Second,
		// WriteTimeout: time.Duration(Conf.TimeoutGin) * time.Second,
		// IdleTimeout:       time.Duration(Conf.TimeoutGin) * time.Second,
	}

	s.srvTLS = &http.Server{
		Addr:        fmt.Sprintf("%s:%s", cfg.Host, cfg.PortTLS),
		Handler:     handler,
		ReadTimeout: cfg.Timeout,
		// ReadHeaderTimeout: time.Duration(Conf.TimeoutGin) * time.Second,
		// WriteTimeout: time.Duration(Conf.TimeoutGin) * time.Second,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS11,
			PreferServerCipherSuites: true,
			SessionTicketsDisabled:   true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,

				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			},
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.CurveP384,
				tls.CurveP521,
			},
		},
	}

	return &s, nil
}

func (s *Server) Start(ctx context.Context) error {
	chErr := make(chan error, 1)

	go func() {
		eg := errgroup.Group{}

		eg.Go(func() error {
			// time.Sleep(5 * time.Second)
			s.log.Infof("Start on [Port: %s]", s.cfg.Port)
			return s.srv.ListenAndServe()
		})

		eg.Go(func() error {
			// time.Sleep(3 * time.Second)
			s.log.Infof("Start on [Port: %s]", s.cfg.PortTLS)
			return s.srvTLS.ListenAndServeTLS(s.cfg.CertCrt, s.cfg.CertKey)
		})

		chErr <- eg.Wait()
	}()

	var err error

	select {
	case <-ctx.Done():
		s.log.Tracef("Server stop: %s", ctx.Err())
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err = s.srv.Shutdown(ctx)
	case err = <-chErr:
		s.log.Errorf("Server error: %s", err.Error())
	}

	return err
}
