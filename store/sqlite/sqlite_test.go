package sqlite

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
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
