package aeat

import (
	"net/http"
	"time"

	"github.com/cristianemek/go-verifactu"
)

const userAgentPorDefecto = "go-verifactu/" + verifactu.Version

// Config represents the configuration for the AEAT client.
type Config struct {
	Entorno         Entorno
	TipoCertificado TipoCertificado
	UserAgent       string
	Timeout         time.Duration
	HTTPClient      *http.Client
}

type Client struct {
	http      *http.Client
	url       string
	userAgent string
}

func NewClient(cfg Config) (*Client, error) {
	url, err := endpoint(cfg.Entorno, cfg.TipoCertificado)

	if err != nil {
		return nil, err
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = userAgentPorDefecto
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	return &Client{
		http:      httpClient,
		url:       url,
		userAgent: cfg.UserAgent,
	}, nil
}
