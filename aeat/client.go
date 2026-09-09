package aeat

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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
	// requires except you bring your own http.Client
	Certificado tls.Certificate
	UserAgent   string
	Timeout     time.Duration
	HTTPClient  *http.Client
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
		if len(cfg.Certificado.Certificate) == 0 {
			return nil, ErrCertificadoRequerido
		}

		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cfg.Certificado},
			MinVersion:   tls.VersionTLS12,
		}

		httpClient = &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
			},
		}
	}

	client := *httpClient

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &Client{
		http:      &client,
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

	// aeat redirects to an error page when the client certificate is missing or not accepted, so we need to handle 3xx status codes explicitly, normally 302
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return record.RespuestaRegFactuSistemaFacturacion{}, fmt.Errorf("%w: status code %d, location: %s", ErrAccesoDenegado, resp.StatusCode, resp.Header.Get("Location"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, err
	}

	respuesta, err := parsearRespuesta(body)

	if err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, fmt.Errorf("%w (http %d)", err, resp.StatusCode)

	}

	return respuesta, nil
}

// ProbarConexion checks that the certificate works against the AEAT without
// filing anything: it sends an empty batch and expects a client fault back.
// It is a real request, so do not call it in a loop or a health check.
func (c *Client) ProbarConexion(ctx context.Context) error {
	_, err := c.Remitir(ctx, verifactu.Tenant{}, record.RegFactuSistemaFacturacion{})

	switch {
	case errors.Is(err, verifactu.ErrFaultCliente):
		return nil
	case err == nil:
		return fmt.Errorf("error inesperado: la AEAT respondió sin error a un lote vacío")
	default:
		return err
	}

}
