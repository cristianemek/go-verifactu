package verifactu_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/memory"
)

func validRegistroAlta(numSerie string) record.RegistroAlta {
	return record.RegistroAlta{
		IDFactura: record.IDFacturaExpedida{
			IDEmisorFactura:        "89890001K",
			NumSerieFactura:        numSerie,
			FechaExpedicionFactura: record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		NombreRazonEmisor:    "EMPRESA DE PRUEBAS SL",
		TipoFactura:          record.TipoFacturaCompleta,
		DescripcionOperacion: "Servicios de desarrollo de software",
		Desglose: record.Desglose{
			DetalleDesglose: []record.DetalleDesglose{
				{
					Impuesto:                      record.Ptr(record.ImpuestoIVA),
					ClaveRegimen:                  record.Ptr(record.ClaveRegimenGeneral),
					CalificacionOperacion:         record.Ptr(record.CalificacionOperacionSujetaNoExentaSinISP),
					TipoImpositivo:                record.Ptr(record.Porcentaje(2100)),
					BaseImponibleOimporteNoSujeto: 10000,
					CuotaRepercutida:              record.Ptr(record.Amount(2100)),
				},
			},
		},
		CuotaTotal:   2100,
		ImporteTotal: 12100,
		SistemaInformatico: record.SistemaInformatico{
			NombreRazon:                 "MI EMPRESA SL",
			NIF:                         record.Ptr("B87654321"),
			NombreSistemaInformatico:    "go-verifactu",
			IdSistemaInformatico:        "01",
			Version:                     "0.1.0",
			NumeroInstalacion:           "0001",
			TipoUsoPosibleSoloVerifactu: record.SiNoSi,
			TipoUsoPosibleMultiOT:       record.SiNoNo,
			IndicadorMultiplesOT:        record.SiNoNo,
		},
	}
}

func validRegistroAnulacion(numSerie string) record.RegistroAnulacion {
	return record.RegistroAnulacion{
		IDFactura: record.IDFacturaExpedidaBaja{
			IDEmisorFacturaAnulada:        "89890001K",
			NumSerieFacturaAnulada:        numSerie,
			FechaExpedicionFacturaAnulada: record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		SistemaInformatico: record.SistemaInformatico{
			NombreRazon:                 "MI EMPRESA SL",
			NIF:                         record.Ptr("B87654321"),
			NombreSistemaInformatico:    "go-verifactu",
			IdSistemaInformatico:        "01",
			Version:                     "0.1.0",
			NumeroInstalacion:           "0001",
			TipoUsoPosibleSoloVerifactu: record.SiNoSi,
			TipoUsoPosibleMultiOT:       record.SiNoNo,
			IndicadorMultiplesOT:        record.SiNoNo,
		},
	}
}

func fixedTime() time.Time {
	return time.Date(2023, 1, 1, 0, 30, 0, 0, time.FixedZone("CET", 3600))
}

func TestAltaPrimerRegistro(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}
	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	entry, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error creating entry: %v", err)
	}

	if entry.Secuencia != 1 {
		t.Fatalf("Expected sequence 1, got %d", entry.Secuencia)
	}

	if entry.Alta.Encadenamiento.PrimerRegistro == nil {
		t.Fatalf("Expected PrimerRegistro to not be nil, got %v", entry.Alta.Encadenamiento.PrimerRegistro)
	}

	if entry.Alta.Encadenamiento.RegistroAnterior != nil {
		t.Fatalf("Expected RegistroAnterior to be nil, got %v", entry.Alta.Encadenamiento.RegistroAnterior)
	}

	if len(entry.Huella) != 64 {
		t.Fatalf("Expected Huella to have 64 characters, got %d", len(entry.Huella))
	}
}

func TestCadenaDeRegistro(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}
	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}
	entry1, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error creating first entry: %v", err)
	}

	entry2, err := engine.Alta(context.Background(), tenant, validRegistroAlta("002"))
	if err != nil {
		t.Fatalf("Error creating second entry: %v", err)
	}

	if entry2.Secuencia != 2 {
		t.Fatalf("Expected sequence 2, got %d", entry2.Secuencia)
	}

	if entry2.Alta.Encadenamiento.PrimerRegistro != nil {
		t.Fatalf("Expected PrimerRegistro to be nil, got %v", entry2.Alta.Encadenamiento.PrimerRegistro)
	}

	if entry2.Alta.Encadenamiento.RegistroAnterior == nil {
		t.Fatalf("Expected RegistroAnterior to not be nil, got %v", entry2.Alta.Encadenamiento.RegistroAnterior)
	}

	if entry2.Alta.Encadenamiento.RegistroAnterior.Huella != entry1.Huella {
		t.Fatalf("Expected RegistroAnterior Huella to match first entry Huella, got %s and %s", entry2.Alta.Encadenamiento.RegistroAnterior.Huella, entry1.Huella)
	}

	if entry2.Huella == entry1.Huella {
		t.Fatalf("Expected different Huella for second entry, got same as first entry: %s", entry2.Huella)
	}

	if entry2.Alta.Encadenamiento.RegistroAnterior.IDEmisorFactura != entry1.Alta.IDFactura.IDEmisorFactura ||
		entry2.Alta.Encadenamiento.RegistroAnterior.NumSerieFactura != entry1.Alta.IDFactura.NumSerieFactura ||
		entry2.Alta.Encadenamiento.RegistroAnterior.FechaExpedicionFactura.Format() != entry1.Alta.IDFactura.FechaExpedicionFactura.Format() {
		t.Fatalf("Expected RegistroAnterior IDFactura to match first entry IDFactura, got %+v and %+v", entry2.Alta.Encadenamiento.RegistroAnterior, entry1.Alta.IDFactura)
	}

}

func TestIdempotenciaAlta(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}
	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}
	entry1, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error creating first entry: %v", err)
	}
	entry2, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error creating second entry: %v", err)
	}

	if entry2.Secuencia != 1 || entry2.Huella != entry1.Huella {
		t.Fatalf("Expected sequence 1 and same Huella for second entry, got %d and %s", entry2.Secuencia, entry2.Huella)
	}

	lastEntry, err := store.Ultimo(context.Background(), tenant)

	if err != nil {
		t.Fatalf("Error retrieving last entry: %v", err)
	}

	if lastEntry.Secuencia != 1 || lastEntry.Huella != entry1.Huella {
		t.Fatalf("Expected last entry to have sequence 1 and same Huella as first entry, got %d and %s", lastEntry.Secuencia, lastEntry.Huella)
	}

}

func TestAltaInvalida(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}
	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}
	invalidRegistro := validRegistroAlta("001")
	invalidRegistro.DescripcionOperacion = ""

	_, err = engine.Alta(context.Background(), tenant, invalidRegistro)
	if !errors.Is(err, record.ErrValidation) {
		t.Fatalf("Expected ErrValidation, got %v", err)
	}

	_, err = store.Ultimo(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Expected ErrNoEncontrado, got %v", err)
	}
}

func TestAislamientoEntreTenants(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	tenant1 := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}
	tenant2 := verifactu.Tenant{NIF: "89890002L", IDSistemaInformatico: "01"}

	_, err = engine.Alta(context.Background(), tenant1, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error creating entry for tenant1: %v", err)
	}

	entry2, err := engine.Alta(context.Background(), tenant1, validRegistroAlta("002"))
	if err != nil {
		t.Fatalf("Error creating entry for tenant1: %v", err)
	}

	entry3, err := engine.Alta(context.Background(), tenant2, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error creating entry for tenant2: %v", err)
	}

	if entry2.Secuencia != 2 || entry2.Alta.Encadenamiento.RegistroAnterior == nil {
		t.Fatalf("Expected sequence 2 and non-nil previous record for entry2, got %d and %v", entry2.Secuencia, entry2.Alta.Encadenamiento.RegistroAnterior)
	}

	if entry3.Secuencia != 1 || entry3.Alta.Encadenamiento.PrimerRegistro == nil {
		t.Fatalf("Expected sequence 1 and non-nil primer record for entry3, got %d and %v", entry3.Secuencia, entry3.Alta.Encadenamiento.PrimerRegistro)
	}

	lastEntry1, err := store.Ultimo(context.Background(), tenant1)
	if err != nil {
		t.Fatalf("Error retrieving last entry for tenant1: %v", err)
	}

	lastEntry2, err := store.Ultimo(context.Background(), tenant2)
	if err != nil {
		t.Fatalf("Error retrieving last entry for tenant2: %v", err)
	}

	if lastEntry1.Secuencia != 2 {
		t.Fatalf("Expected last entry for tenant1 to have sequence 2, got %d", lastEntry1.Secuencia)
	}

	if lastEntry2.Secuencia != 1 {
		t.Fatalf("Expected last entry for tenant2 to have sequence 1, got %d", lastEntry2.Secuencia)
	}

}

const (
	numAltas uint64 = 50
)

func TestConcurrencia(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	entries := make([]*verifactu.Entry, numAltas)
	errorsCh := make(chan error, numAltas)

	var wg sync.WaitGroup

	for i := range numAltas {
		wg.Add(1)
		go func() {
			defer wg.Done()
			entry, err := engine.Alta(context.Background(), tenant, validRegistroAlta(fmt.Sprintf("03%d", i+1)))
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

	lastEntry, err := store.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Error retrieving last entry: %v", err)
	}

	if lastEntry.Secuencia != numAltas {
		t.Fatalf("Expected last entry to have sequence %d, got %d", numAltas, lastEntry.Secuencia)
	}

	porSecuencia := make([]*verifactu.Entry, numAltas+1)

	for i := range entries {
		if entries[i].Secuencia < 1 || entries[i].Secuencia > numAltas {
			t.Fatalf("Entry %d has invalid sequence number %d", i+1, entries[i].Secuencia)
		}

		if porSecuencia[entries[i].Secuencia] != nil {
			t.Fatalf("Duplicate sequence number %d found", entries[i].Secuencia)
		}

		porSecuencia[entries[i].Secuencia] = entries[i]

	}

	for i := uint64(1); i <= numAltas; i++ {
		if porSecuencia[i] == nil {
			t.Fatalf("Missing entry for sequence number %d", i)
		}
	}

	if porSecuencia[1].Alta.Encadenamiento.PrimerRegistro == nil {
		t.Fatalf("Expected PrimerRegistro to be non-nil for first entry, got nil")
	}

	if porSecuencia[1].Alta.Encadenamiento.RegistroAnterior != nil {
		t.Fatalf("Expected RegistroAnterior to be nil for first entry, got %v", porSecuencia[1].Alta.Encadenamiento.RegistroAnterior)
	}

	for i := uint64(2); i <= numAltas; i++ {
		actual := porSecuencia[i]
		anterior := porSecuencia[i-1]

		if actual.Alta.Encadenamiento.RegistroAnterior == nil {
			t.Fatalf("Expected RegistroAnterior to be non-nil for entry %d, got nil", i)
		}

		if actual.Alta.Encadenamiento.RegistroAnterior.Huella != anterior.Huella {
			t.Fatalf("Expected RegistroAnterior Huella to match previous entry Huella for entry %d, got %s and %s", i, actual.Alta.Encadenamiento.RegistroAnterior.Huella, anterior.Huella)
		}
	}

}

func TestAltaAnulacion(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}
	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	entryAlta, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))
	if err != nil {
		t.Fatalf("Error creating alta entry: %v", err)
	}

	entryAnulacion, err := engine.Anular(context.Background(), tenant, validRegistroAnulacion("001"))
	if err != nil {
		t.Fatalf("Error creating anulacion entry: %v", err)
	}

	if entryAnulacion.Secuencia != 2 {
		t.Fatalf("Expected sequence 2 for anulacion entry, got %d", entryAnulacion.Secuencia)
	}

	if entryAnulacion.Operacion != verifactu.OperacionAnulacion || entryAnulacion.Anulacion == nil || entryAnulacion.Alta != nil {
		t.Fatalf("Expected operation 'anulacion' for anulacion entry, got %s", entryAnulacion.Operacion)
	}

	if entryAnulacion.Anulacion.Encadenamiento.RegistroAnterior == nil || entryAnulacion.Anulacion.Encadenamiento.RegistroAnterior.Huella != entryAlta.Huella {
		t.Fatalf("Expected RegistroAnterior to be non-nil for anulacion entry, got nil")
	}

	lastEntry, err := store.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Error retrieving last entry: %v", err)
	}

	if lastEntry.Secuencia != 2 {
		t.Fatalf("Expected last entry to have sequence 2, got %d", lastEntry.Secuencia)
	}

	gotEntryAlta, err := store.Buscar(context.Background(), tenant, entryAlta.IDFactura, verifactu.OperacionAlta)

	if err != nil {
		t.Fatalf("Error retrieving alta entry: %v", err)
	}

	if gotEntryAlta.Secuencia != entryAlta.Secuencia {
		t.Fatalf("Expected alta entry with Secuencia %v, got %v", entryAlta.Secuencia, gotEntryAlta.Secuencia)
	}

	gotEntryAnulacion, err := store.Buscar(context.Background(), tenant, entryAnulacion.IDFactura, verifactu.OperacionAnulacion)

	if err != nil {
		t.Fatalf("Error retrieving anulacion entry: %v", err)
	}

	if gotEntryAnulacion.Secuencia != entryAnulacion.Secuencia {
		t.Fatalf("Expected anulacion entry with Secuencia %v, got %v", entryAnulacion.Secuencia, gotEntryAnulacion.Secuencia)
	}

}

func TestEstado(t *testing.T) {
	store := memory.New()
	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}
	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	entryAlta, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))

	if err != nil {
		t.Fatalf("Error creating alta entry: %v", err)
	}

	testCases := []struct {
		name      string
		id        verifactu.IDFactura
		op        verifactu.Operacion
		wantEntry *verifactu.Entry
		wantErr   error
	}{
		{
			name:      "Estado de alta existente",
			id:        entryAlta.IDFactura,
			op:        verifactu.OperacionAlta,
			wantEntry: entryAlta,
			wantErr:   nil,
		},
		{
			name:      "Diferente operación para alta existente",
			id:        entryAlta.IDFactura,
			op:        verifactu.OperacionAnulacion,
			wantEntry: nil,
			wantErr:   verifactu.ErrNoEncontrado,
		},
		{
			name:      "Numero de serie inexistente",
			id:        verifactu.IDFactura{NIF: "89890001K", NumSerie: "002", Fecha: record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))},
			op:        verifactu.OperacionAlta,
			wantEntry: nil,
			wantErr:   verifactu.ErrNoEncontrado,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotEntry, gotErr := engine.Estado(context.Background(), tenant, tc.id, tc.op)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("Expected error %v, got %v", tc.wantErr, gotErr)
			}

			if tc.wantEntry != gotEntry {
				t.Errorf("Expected entry %v, got %v", tc.wantEntry, gotEntry)
			}

		})

	}
}

func TestEntryJson(t *testing.T) {
	store := memory.New()

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	entry, err := engine.Alta(context.Background(), verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}, validRegistroAlta("001"))

	if err != nil {
		t.Fatalf("Error creating alta entry: %v", err)
	}

	entryJson, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Error marshaling entry to JSON: %v", err)
	}

	var recuperada verifactu.Entry

	err = json.Unmarshal(entryJson, &recuperada)
	if err != nil {
		t.Fatalf("Error unmarshaling JSON to entry: %v", err)
	}

	if recuperada.Huella != entry.Huella {
		t.Fatalf("Expected Huella %s, got %s", entry.Huella, recuperada.Huella)
	}

	if recuperada.Alta.Fingerprint() != recuperada.Huella {
		t.Fatalf("Expected Alta fingerprint %s, got %s", entry.Alta.Fingerprint(), recuperada.Alta.Fingerprint())
	}

}
