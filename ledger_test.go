package verifactu_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
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

func TestLedgerEnviosSobrevivenReinicio(t *testing.T) {
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
	_, err = engine.Alta(context.Background(), tenant, validRegistroAlta("002"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}
	entry3, err := engine.Alta(context.Background(), tenant, validRegistroAlta("003"))
	if err != nil {
		t.Fatalf("Error calling Alta: %v", err)
	}

	err = store.AnexarEnvio(context.Background(), tenant, &verifactu.Envio{
		CSV: "CSV1",
		Lineas: []verifactu.LineaEnvio{
			{Secuencia: entry1.Secuencia, Estado: record.EstadoRegistroCorrecto},
			{Secuencia: entry3.Secuencia, Estado: record.EstadoRegistroCorrecto},
		},
	})

	if err != nil {
		t.Fatalf("Error calling AnexarEnvio: %v", err)
	}

	pendientes, err := store.Pendientes(context.Background(), tenant, 0)

	if err != nil {
		t.Fatalf("Error calling Pendientes: %v", err)
	}

	if len(pendientes) != 1 {
		t.Fatalf("Expected 1 pending entries, got %d", len(pendientes))
	}

	if pendientes[0].Secuencia != 2 {
		t.Fatalf("Expected pending entry to be entry2, got %+v", pendientes[0])
	}

	store, err = ledger.New(dir) //Simulates a restart

	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	pendientes, err = store.Pendientes(context.Background(), tenant, 0)
	if err != nil {
		t.Fatalf("Error calling Pendientes: %v", err)
	}

	if len(pendientes) != 1 {
		t.Fatalf("Expected 1 pending entries, got %d", len(pendientes))
	}

	if pendientes[0].Secuencia != 2 {
		t.Fatalf("Expected pending entry to be entry2, got %+v", pendientes[0])
	}

	ultimoEnvio, err := store.UltimoEnvio(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Error calling UltimoEnvio: %v", err)
	}

	if ultimoEnvio.CSV != "CSV1" {
		t.Fatalf("Expected ultimo envio CSV to be 'CSV1', got '%s'", ultimoEnvio.CSV)
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

func TestLedgerConcurrenciaMantieneElOrden(t *testing.T) {
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

	entries := make([]*verifactu.Entry, numAltas)
	errorsCh := make(chan error, numAltas)

	var wg sync.WaitGroup

	for i := range numAltas {
		wg.Add(1)
		go func() {
			defer wg.Done()
			entry, err := engine.Alta(context.Background(), tenant, validRegistroAlta(fmt.Sprintf("%03d", i+1)))
			if err != nil {
				errorsCh <- fmt.Errorf("Error creating entry %d: %v", i+1, err)
				return
			}
			entries[i] = entry
		}()
	}
	wg.Wait()
	close(errorsCh)

	for err := range errorsCh {
		t.Errorf("Error creating entry: %v", err)
	}

	if t.Failed() {
		t.FailNow()
	}

	store, err = ledger.New(dir) //Simulates a restart
	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	ultimoEntry, err := store.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Error calling Ultimo: %v", err)
	}

	if ultimoEntry.Secuencia != numAltas {
		t.Fatalf("Expected last entry sequence to be %d, got %d", numAltas, ultimoEntry.Secuencia)
	}

	ruta := filepath.Join(dir, "89890001K-01.jsonl")

	fileContent, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	fileContentStr := strings.TrimRight(string(fileContent), "\n")
	fileLines := strings.Split(fileContentStr, "\n")

	if len(fileLines) != int(numAltas) {
		t.Fatalf("Expected %d lines in the file, got %d", numAltas, len(fileLines))
	}

	previousEntry := verifactu.Entry{}

	for i, line := range fileLines {
		var entry verifactu.Entry
		err := json.Unmarshal([]byte(line), &entry)
		if err != nil {
			t.Fatalf("Error unmarshalling line %d: %v", i+1, err)
		}
		if entry.Secuencia != uint64(i+1) {
			t.Fatalf("Expected entry sequence to be %d, got %d", i+1, entry.Secuencia)
		}

		if i == 0 && entry.Alta.Encadenamiento.PrimerRegistro == nil {
			t.Fatalf("Expected first entry to have PrimerRegistro, got nil")
		}

		if i > 0 {
			if entry.Alta.Encadenamiento.RegistroAnterior == nil {
				t.Fatalf("Expected entry %d to have RegistroAnterior, got nil", i+1)
			}
			if entry.Alta.Encadenamiento.RegistroAnterior.Huella != previousEntry.Huella {
				t.Fatalf("Expected entry %d to have correct RegistroAnterior, got incorrect", i+1)
			}
		}

		previousEntry = entry
	}

}
