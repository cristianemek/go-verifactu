package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/aeat"
	"github.com/cristianemek/go-verifactu/store/ledger"
)

func main() {
	path := flag.String("config", "verifactud.json", "Path to the configuration file")

	flag.Parse()

	cfg, err := cargarConfig(*path)
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		os.Exit(1)
	}

	srv, err := construirServidor(cfg)
	if err != nil {
		slog.Error("Error constructing server", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", srv.healthz)
	mux.HandleFunc("POST /v1/{nif}/alta", srv.auth(srv.alta))
	mux.HandleFunc("POST /v1/{nif}/anular", srv.auth(srv.anulacion))
	mux.HandleFunc("GET /v1/{nif}/estado", srv.auth(srv.estado))

	slog.Info("Listening", "address", cfg.Listen)

	err = http.ListenAndServe(cfg.Listen, mux)

	if err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}

func construirServidor(cfg *Config) (*servidor, error) {
	store, err := ledger.New(cfg.Data)

	if err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}

	cert, err := aeat.CargarPEM(cfg.Certificado, cfg.Certificado)

	if err != nil {
		return nil, fmt.Errorf("error loading certificate: %w", err)
	}

	cliente, err := aeat.NewClient(aeat.Config{
		Entorno:         aeat.Entorno(cfg.Entorno),
		TipoCertificado: aeat.TipoCertificado(cfg.TipoCertificado),
		Certificado:     cert,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	engine, err := verifactu.New(verifactu.Config{
		Store:              store,
		Transport:          cliente,
		SistemaInformatico: &cfg.Sistema,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating engine: %w", err)
	}

	return &servidor{
		engine:  engine,
		cliente: cliente,
		tenants: cfg.Tenants,
		sistema: cfg.Sistema.IdSistemaInformatico,
	}, nil
}
