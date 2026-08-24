package verifactu_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/store/ledger"
)

func TestLedgerSobreviveReinicio(t *testing.T) {
	dir := t.TempDir()

	store, err := ledger.New(dir)
	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating verifactu engine: %v", err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	entry1, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}
	entry2, err := engine.Alta(context.Background(), tenant, validRegistroAlta("002"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}
	entry3, err := engine.Alta(context.Background(), tenant, validRegistroAlta("003"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}

	store, err = ledger.New(dir) //Simulates a restart
	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	lastEntry, err := store.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Error calling Ultimo: %v", err)
	}

	if lastEntry.Secuencia != 3 || lastEntry.Huella != entry3.Huella {
		t.Errorf("Expected last entry to be entry3, got %+v", lastEntry)
	}

	searchEntry, err := store.Buscar(context.Background(), tenant, entry1.IDFactura, verifactu.OperacionAlta)
	if err != nil {
		t.Fatalf("Error calling Buscar: %v", err)
	}

	if searchEntry.Secuencia != 1 || searchEntry.Huella != entry1.Huella {
		t.Errorf("Expected search entry to be entry1, got %+v", searchEntry)
	}

	searchEntry, err = store.Buscar(context.Background(), tenant, entry2.IDFactura, verifactu.OperacionAlta)
	if err != nil {
		t.Fatalf("Error calling Buscar: %v", err)
	}

	if searchEntry.Secuencia != 2 || searchEntry.Huella != entry2.Huella {
		t.Errorf("Expected search entry to be entry2, got %+v", searchEntry)
	}

	if lastEntry.Alta.Fingerprint() != lastEntry.Huella {
		t.Errorf("Expected last entry fingerprint to match its huella, got %s and %s", lastEntry.Alta.Fingerprint(), lastEntry.Huella)
	}

	engine, err = verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error recreating verifactu engine: %v", err)
	}

	entry4, err := engine.Alta(context.Background(), tenant, validRegistroAlta("004"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}

	if entry4.Secuencia != 4 {
		t.Errorf("Expected entry4 sequence to be 4, got %d", entry4.Secuencia)
	}

	if entry4.Alta.Encadenamiento.RegistroAnterior == nil {
		t.Fatalf("Expected entry4 to have a previous registro, got %+v", entry4)
	}

	if entry4.Alta.Encadenamiento.RegistroAnterior.Huella != entry3.Huella {
		t.Errorf("Expected entry4 previous fingerprint to match entry3 fingerprint, got %s and %s", entry4.Alta.Encadenamiento.RegistroAnterior.Huella, entry3.Huella)
	}

}

func TestLedgerToleraLineaIncompleta(t *testing.T) {
	dir := t.TempDir()

	store, err := ledger.New(dir)
	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating verifactu engine: %v", err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	_, err = engine.Alta(context.Background(), tenant, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}
	_, err = engine.Alta(context.Background(), tenant, validRegistroAlta("002"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}
	entry3, err := engine.Alta(context.Background(), tenant, validRegistroAlta("003"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}

	ruta := filepath.Join(dir, "89890001K-01.jsonl")

	file, err := os.OpenFile(ruta, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("Error opening file: %v", err)
	}

	_, err = file.WriteString(`{"Operacion":"alta","Sec`)
	if err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	file.Close()

	store, err = ledger.New(dir)
	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	ultimoEntry, err := store.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Error calling Ultimo: %v", err)
	}

	if ultimoEntry.Secuencia != 3 {
		t.Fatalf("Expected last entry sequence to be 3, got %d", ultimoEntry.Secuencia)
	}

	engine, err = verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating verifactu engine: %v", err)
	}

	entry4, err := engine.Alta(context.Background(), tenant, validRegistroAlta("004"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}

	if entry4.Alta.Encadenamiento.RegistroAnterior == nil {
		t.Fatalf("Expected entry4 to have a encadenamiento RegistroAnterior, got nil")
	}

	if entry4.Secuencia != 4 || entry4.Alta.Encadenamiento.RegistroAnterior.Huella != entry3.Huella {
		t.Fatalf("Expected entry4 sequence to be 4 and previous fingerprint to match entry3, got sequence %d and previous fingerprint %s", entry4.Secuencia, entry4.Alta.Encadenamiento.RegistroAnterior.Huella)
	}

	store, err = ledger.New(dir)
	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	ultimoEntry, err = store.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Error calling Ultimo: %v", err)
	}

	if ultimoEntry.Secuencia != 4 {
		t.Fatalf("Expected last entry sequence to be 4, got %d", ultimoEntry.Secuencia)
	}

}
