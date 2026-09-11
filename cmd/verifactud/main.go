package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	mux.HandleFunc("GET /v1/{nif}/conexion", srv.auth(srv.conexion))

	cada, err := time.ParseDuration(cfg.RemisionCada)

	if err != nil {
		slog.Error("Error parsing remision interval", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	hecho := make(chan struct{})

	go func() {
		srv.bucleRemision(ctx, cada)
		close(hecho)
	}()

	httpSrv := &http.Server{Addr: cfg.Listen, Handler: mux}

	slog.Info("Listening", "address", cfg.Listen)

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("Shutting down server")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = httpSrv.Shutdown(ctxShutdown)

	if err != nil {
		slog.Error("Error shutting down server", "error", err)
	}
	<-hecho

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
