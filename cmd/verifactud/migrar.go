package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/store/ledger"
	"github.com/cristianemek/go-verifactu/store/sqlite"
)

func comandoMigrar(args []string) error {
	fs := flag.NewFlagSet("migrar", flag.ExitOnError)

	config := fs.String("config", "verifactud.json", "Path to the configuration file")

	destino := fs.String("a", "", "Path to the target SQLite database file")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if *destino == "" {
		return fmt.Errorf("-a is required: path to the target SQLite database file")
	}

	cfg, err := cargarConfig(*config)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	slog.Info("Starting migration", "target", *destino, "origen", cfg.Data, "tenants", len(cfg.Tenants))

	if cfg.Store != StoreLedger {
		return fmt.Errorf("migration is only supported from ledger store, current store: %s", cfg.Store)
	}

	if _, err := os.Stat(*destino); err == nil {
		return fmt.Errorf("target SQLite database file already exists: %s", *destino)
	}

	store, err := ledger.New(cfg.Data)

	if err != nil {
		return fmt.Errorf("error creating ledger store: %w", err)
	}

	sqliteStore, err := sqlite.New(*destino)

	if err != nil {
		return fmt.Errorf("error creating sqlite store: %w", err)
	}

	defer sqliteStore.Close()

	for nif := range cfg.Tenants {
		tenant := verifactu.Tenant{NIF: nif, IDSistemaInformatico: cfg.Sistema.IdSistemaInformatico}

		entries, err := store.Cadena(context.Background(), tenant)
		if err != nil {
			return fmt.Errorf("error migrating tenant %s: %w", nif, err)
		}

		slog.Info("Migrating tenant", "nif", nif, "entries", len(entries))

		for _, entry := range entries {
			err := sqliteStore.Anexar(context.Background(), tenant, entry)
			if err != nil {
				return fmt.Errorf("error migrating entry for tenant %s: %w, secuencia: %d", nif, err, entry.Secuencia)
			}
		}

		migrado := map[time.Time]bool{}
		contador := 0

		for _, entry := range entries {
			envio, err := store.EnvioDe(context.Background(), tenant, entry.Secuencia)

			if errors.Is(err, verifactu.ErrNoEncontrado) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error retrieving envio for tenant %s, secuencia %d: %w", nif, entry.Secuencia, err)
			}

			if _, ok := migrado[envio.Instante]; ok {
				continue
			}

			err = sqliteStore.AnexarEnvio(context.Background(), tenant, envio)

			if err != nil {
				return fmt.Errorf("error migrating envio for tenant %s, secuencia %d: %w", nif, entry.Secuencia, err)
			}

			migrado[envio.Instante] = true
			contador++
		}

		sqliteEntries, err := sqliteStore.Cadena(context.Background(), tenant)
		if err != nil {
			return fmt.Errorf("error retrieving cadena for tenant %s: %w", nif, err)
		}

		err = verifactu.VerificarCadena(sqliteEntries)

		if err != nil {
			return fmt.Errorf("error verifying cadena for tenant %s: %w", nif, err)
		}

		ledgerPendientes, err := store.Pendientes(context.Background(), tenant, 0)
		if err != nil {
			return fmt.Errorf("error retrieving pendientes for tenant %s: %w", nif, err)
		}

		sqlitePendientes, err := sqliteStore.Pendientes(context.Background(), tenant, 0)
		if err != nil {
			return fmt.Errorf("error retrieving pendientes for tenant %s: %w", nif, err)
		}

		if len(ledgerPendientes) != len(sqlitePendientes) {
			return fmt.Errorf("pendientes count mismatch for tenant %s: ledger=%d, sqlite=%d", nif, len(ledgerPendientes), len(sqlitePendientes))
		}

		for i := range ledgerPendientes {
			if ledgerPendientes[i].Secuencia != sqlitePendientes[i].Secuencia {
				return fmt.Errorf("pendientes mismatch for tenant %s at index %d: ledger=%d, sqlite=%d", nif, i, ledgerPendientes[i].Secuencia, sqlitePendientes[i].Secuencia)
			}
		}

		ledgerUltimoEnvio, err := store.UltimoEnvio(context.Background(), tenant)

		var origenSinEnvios bool

		if err != nil {
			if errors.Is(err, verifactu.ErrNoEncontrado) {
				origenSinEnvios = true
			}

			return fmt.Errorf("error retrieving ultimo envio for tenant %s: %w", nif, err)
		}

		var destinoSinEnvios bool

		sqliteUltimoEnvio, err := sqliteStore.UltimoEnvio(context.Background(), tenant)
		if err != nil {
			if errors.Is(err, verifactu.ErrNoEncontrado) {
				destinoSinEnvios = true
			}

			return fmt.Errorf("error retrieving ultimo envio for tenant %s: %w", nif, err)
		}

		if origenSinEnvios != destinoSinEnvios {
			return fmt.Errorf("ultimo envio presence mismatch for tenant %s: origenSinEnvios=%v, destinoSinEnvios=%v", nif, origenSinEnvios, destinoSinEnvios)
		}

		if !origenSinEnvios && !destinoSinEnvios {

			if !ledgerUltimoEnvio.Instante.Equal(sqliteUltimoEnvio.Instante) {
				return fmt.Errorf("ultimo envio mismatch for tenant %s: ledger=%v, sqlite=%v", nif, ledgerUltimoEnvio.Instante, sqliteUltimoEnvio.Instante)
			}

			if ledgerUltimoEnvio.CSV != sqliteUltimoEnvio.CSV {
				return fmt.Errorf("ultimo envio CSV mismatch for tenant %s: ledger=%s, sqlite=%s", nif, ledgerUltimoEnvio.CSV, sqliteUltimoEnvio.CSV)
			}
		}

		slog.Info("Finished migrating tenant", "nif", nif, "entries", len(entries), "envios", contador)

	}

	return nil
}
