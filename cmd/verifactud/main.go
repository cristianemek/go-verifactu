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
	"github.com/cristianemek/go-verifactu/store/sqlite"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := ejecutar(ctx, os.Args[1:])

	stop()

	if err != nil {
		slog.Error("verifactud", "error", err)
		os.Exit(1)
	}
}

func ejecutar(ctx context.Context, args []string) error {
	if len(args) > 0 && args[0] == "migrar" {
		if err := comandoMigrar(args[1:]); err != nil {
			return fmt.Errorf("error migrating data: %w", err)
		}
		slog.Info("Data migrated successfully")
		return nil
	}

	fs := flag.NewFlagSet("verifactud", flag.ContinueOnError)

	path := fs.String("config", "verifactud.json", "Path to the configuration file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := cargarConfig(*path)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	cada, err := time.ParseDuration(cfg.RemisionCada)
	if err != nil {
		return fmt.Errorf("error parsing remision interval: %w", err)
	}

	if cfg.Log != "" {
		f, err := os.OpenFile(cfg.Log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
		if err != nil {
			return fmt.Errorf("error opening log file: %w", err)
		}
		defer f.Close()

		slog.SetDefault(slog.New(slog.NewJSONHandler(f, nil)))
	}

	srv, err := construirServidor(cfg)
	if err != nil {
		return fmt.Errorf("error constructing server: %w", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", srv.healthz)
	mux.HandleFunc("POST /v1/{nif}/alta", srv.auth(srv.alta))
	mux.HandleFunc("POST /v1/{nif}/anular", srv.auth(srv.anulacion))
	mux.HandleFunc("GET /v1/{nif}/estado", srv.auth(srv.estado))
	mux.HandleFunc("GET /v1/{nif}/conexion", srv.auth(srv.conexion))

	ctx, cancelar := context.WithCancel(ctx)
	defer cancelar()

	hecho := make(chan struct{})

	go func() {
		srv.bucleRemision(ctx, cada)
		close(hecho)
	}()

	httpSrv := &http.Server{Addr: cfg.Listen, Handler: mux}

	slog.Info("Listening", "address", cfg.Listen)

	errServidor := make(chan error, 1)

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errServidor <- fmt.Errorf("error starting server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err = <-errServidor:
	}

	slog.Info("Shutting down server")

	cancelar()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctxShutdown); err != nil {
		slog.Error("Error shutting down server", "error", err)
	}

	<-hecho

	if err := srv.cerrar(); err != nil {
		slog.Error("Error closing store", "error", err)
	}

	return err
}

func construirServidor(cfg *Config) (*servidor, error) {
	var verifactuStore verifactu.Store

	switch cfg.Store {
	case StoreLedger:
		store, err := ledger.New(cfg.Data)
		if err != nil {
			return nil, fmt.Errorf("error creating ledger store: %w", err)
		}
		verifactuStore = store
	case StoreSQLite:
		store, err := sqlite.New(cfg.Data)
		if err != nil {
			return nil, fmt.Errorf("error creating sqlite store: %w", err)
		}
		verifactuStore = store
	}

	clientes := make(map[string]*aeat.Client, len(cfg.Tenants))
	transportes := make(map[string]verifactu.Transport, len(cfg.Tenants))

	for nif, t := range cfg.Tenants {
		cert, err := aeat.CargarPEM(t.Certificado, t.Certificado)
		if err != nil {
			return nil, fmt.Errorf("error loading certificate for tenant %s: %w", nif, err)
		}

		cliente, err := aeat.NewClient(aeat.Config{
			Entorno:         aeat.Entorno(cfg.Entorno),
			TipoCertificado: aeat.TipoCertificado(t.TipoCertificado),
			Certificado:     cert,
		})

		if err != nil {
			return nil, fmt.Errorf("error creating client for tenant %s: %w", nif, err)
		}

		clientes[nif] = cliente
		transportes[nif] = cliente

	}

	engine, err := verifactu.New(verifactu.Config{
		Store:              verifactuStore,
		Transport:          &transportePorTenant{transportes: transportes},
		SistemaInformatico: &cfg.Sistema,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating engine: %w", err)
	}

	for nif := range cfg.Tenants {
		tenant := verifactu.Tenant{NIF: nif, IDSistemaInformatico: cfg.Sistema.IdSistemaInformatico}

		cadena, err := verifactuStore.Cadena(context.Background(), tenant)

		if err != nil {
			return nil, fmt.Errorf("error retrieving chain for tenant %s: %w", nif, err)
		}

		if err := verifactu.VerificarCadena(cadena); err != nil {
			return nil, fmt.Errorf("error verifying chain for tenant %s: %w", nif, err)
		}

	}

	return &servidor{
		engine:   engine,
		clientes: clientes,
		tenants:  cfg.Tenants,
		sistema:  cfg.Sistema.IdSistemaInformatico,
		avisar:   make(chan struct{}, 1),
		entorno:  entornoQR(cfg.Entorno),
		almacen:  verifactuStore,
	}, nil
}
