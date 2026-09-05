package aeat

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
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

var _ verifactu.Transport = (*Client)(nil)

func (c *Client) Remitir(ctx context.Context, t verifactu.Tenant, lote record.RegFactuSistemaFacturacion) (record.RespuestaRegFactuSistemaFacturacion, error) {

	peticion, err := serializarSobre(lote)

	if err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(peticion))
	if err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("SOAPAction", "")

	resp, err := c.http.Do(req)
	if err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, err
	}

	return parsearRespuesta(body)
}
