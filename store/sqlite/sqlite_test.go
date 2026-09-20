package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/storetest"
)

func TestNewReabrirNoAplica(t *testing.T) {
	dir := t.TempDir()

	dbPath := filepath.Join(dir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Error creating store: %v", err)
	}

	row := store.db.QueryRow("PRAGMA user_version")
	var version int
	err = row.Scan(&version)

	if err != nil {
		t.Fatalf("Error querying schema version: %v", err)
	}

	if version != 1 {
		t.Fatalf("Expected schema version 1, got %d", version)
	}

	row = store.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN ('entradas', 'envios', 'lineas')")

	tableCount := 0
	err = row.Scan(&tableCount)

	if err != nil {
		t.Fatalf("Error querying table count: %v", err)
	}

	if tableCount != 3 {
		t.Fatalf("Expected 3 tables, but not all were created")
	}

	store.Close()

	store2, err := New(dbPath)
	if err != nil {
		t.Fatalf("Error creating store: %v", err)
	}

	store2.Close()

}

func TestNewVersionDelFuturo(t *testing.T) {

	dir := t.TempDir()

	dbPath := filepath.Join(dir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Error creating store: %v", err)
	}

	store.Close()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}

	_, err = db.Exec("PRAGMA user_version = 999")
	if err != nil {
		t.Fatalf("Error setting schema version: %v", err)
	}
	db.Close()

	_, err = New(dbPath)
	if err == nil {
		t.Fatalf("Expected error due to future schema version, but got none")
	}

	if !strings.Contains(err.Error(), "999") {
		t.Fatalf("Expected error message to contain '999', but got: %v", err)
	}

}

func TestEjecutarSQL(t *testing.T) {
	dir := t.TempDir()

	dbPath := filepath.Join(dir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Error creating store: %v", err)
	}

	defer store.Close()

	tenant := verifactu.Tenant{
		NIF:                  "89890001K",
		IDSistemaInformatico: "01",
	}

	entry := &verifactu.Entry{
		Secuencia: 1,
		Operacion: verifactu.OperacionAlta,
		IDFactura: verifactu.IDFactura{
			NIF:      "89890001K",
			NumSerie: "A001",
			Fecha:    record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	err = store.Anexar(context.Background(), tenant, entry)
	if err != nil {
		t.Fatalf("Error anexing entry: %v", err)
	}

	entry.Secuencia = 2
	err = store.Anexar(context.Background(), tenant, entry)

	if err == nil {
		t.Fatalf("Expected error due to duplicate entry, but got none")
	}

	if !errors.Is(err, verifactu.ErrDuplicado) {
		t.Fatalf("Expected ErrDuplicado, but got: %v", err)
	}

	entry.Secuencia = 5
	entry.IDFactura.NumSerie = "A002"

	err = store.Anexar(context.Background(), tenant, entry)

	if err == nil {
		t.Fatalf("Expected error due to non-sequential sequence, but got none")
	}

	if !errors.Is(err, verifactu.ErrConflictoDeSecuencia) {
		t.Fatalf("Expected ErrConflictoDeSecuencia, but got: %v", err)
	}

}

func TestSQLiteSobreviveReinicio(t *testing.T) {
	dir := t.TempDir()

	dbPath := filepath.Join(dir, "verifactu.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Error creating store: %v", err)
	}

	tenant := verifactu.Tenant{
		NIF:                  "89890001K",
		IDSistemaInformatico: "01",
	}

	entry1 := &verifactu.Entry{
		Secuencia: 1,
		Operacion: verifactu.OperacionAlta,
		IDFactura: verifactu.IDFactura{
			NIF:      "89890001K",
			NumSerie: "A001",
			Fecha:    record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	entry2 := &verifactu.Entry{
		Secuencia: 2,
		Operacion: verifactu.OperacionAlta,
		IDFactura: verifactu.IDFactura{
			NIF:      "89890001K",
			NumSerie: "A002",
			Fecha:    record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	err = store.Anexar(context.Background(), tenant, entry1)

	if err != nil {
		t.Fatalf("Error anexing entry1: %v", err)
	}

	err = store.Anexar(context.Background(), tenant, entry2)

	if err != nil {
		t.Fatalf("Error anexing entry2: %v", err)
	}

	store.Close()

	store2, err := New(dbPath)

	if err != nil {
		t.Fatalf("Error reopening store: %v", err)
	}

	envio := &verifactu.Envio{
		Instante:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		CSV:            "A-1",
		NIFPresentador: "89890001K",
		EstadoEnvio:    record.EstadoEnvioCorrecto,
		Lineas: []verifactu.LineaEnvio{
			{
				Secuencia: 1,
				Estado:    record.EstadoRegistroCorrecto,
			},
			{
				Secuencia: 2,
				Estado:    record.EstadoRegistroCorrecto,
			},
		},
		TiempoEspera: 60 * time.Second,
	}

	err = store2.AnexarEnvio(context.Background(), tenant, envio)

	if err != nil {
		t.Fatalf("Error anexing envio: %v", err)
	}

	t.Cleanup(func() {
		store2.Close()
	})

	ultimo, err := store2.Ultimo(context.Background(), tenant)

	if err != nil {
		t.Fatalf("Error getting ultimo: %v", err)
	}

	if ultimo.Secuencia != 2 {
		t.Fatalf("Expected ultimo secuencia to be 2, but got: %v", ultimo.Secuencia)
	}

	ultimoEnvio, err := store2.UltimoEnvio(context.Background(), tenant)

	if err != nil {
		t.Fatalf("Error getting ultimo envio: %v", err)
	}

	if ultimoEnvio.CSV != "A-1" {
		t.Fatalf("Expected ultimo envio CSV to be 'A-1', but got: %v", ultimoEnvio)
	}

	if len(ultimoEnvio.Lineas) != 2 {
		t.Fatalf("Expected ultimo envio to have 2 lines, but got: %v", len(ultimoEnvio.Lineas))
	}

	if !ultimoEnvio.Instante.Equal(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("Expected ultimo envio instante to be 2024-01-02, but got: %v", ultimoEnvio.Instante)
	}

	cadena, err := store2.Cadena(context.Background(), tenant)

	if err != nil {
		t.Fatalf("Error getting cadena: %v", err)
	}

	if len(cadena) != 2 {
		t.Fatalf("Expected cadena to be 2, but got: %v", cadena)
	}
}

func TestConformidad(t *testing.T) {

	storetest.Conformidad(t, func(t *testing.T) verifactu.Store {
		t.Helper()
		dir := t.TempDir()

		ruta := filepath.Join(dir, "verifactu.db")
		store, err := New(ruta)
		if err != nil {
			t.Fatalf("Error creating store: %v", err)
		}

		t.Cleanup(func() {
			store.Close()
		})
		return store
	})
}
