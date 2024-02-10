package apiservice

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Address string
	Timeout time.Duration
}

type Client struct {
	address string
	*http.Client
}

func NewClient(cfg Config) *Client {
	dialer := &net.Dialer{
		Timeout:   cfg.Timeout,
		KeepAlive: cfg.Timeout,
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			Dial:              dialer.Dial,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &Client{
		cfg.Address,
		httpClient,
	}
}
