package verifactu

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu/record"
)

type storeFalso struct{}

func (s storeFalso) Ultimo(ctx context.Context, t Tenant) (*Entry, error) {
	return nil, ErrNoEncontrado
}

func (s storeFalso) Anexar(ctx context.Context, t Tenant, e *Entry) error {
	return nil
}

func (s storeFalso) Buscar(ctx context.Context, t Tenant, id IDFactura, op Operacion) (*Entry, error) {
	return nil, ErrNoEncontrado
}

// AnexarEnvio implements [verifactu.Store].
func (s storeFalso) AnexarEnvio(ctx context.Context, t Tenant, envio *Envio) error {
	panic("unimplemented")
}

// Pendientes implements [verifactu.Store].
func (s storeFalso) Pendientes(ctx context.Context, t Tenant, limite int) ([]*Entry, error) {
	panic("unimplemented")
}

// UltimoEnvio implements [verifactu.Store].
func (s storeFalso) UltimoEnvio(ctx context.Context, t Tenant) (*Envio, error) {
	panic("unimplemented")
}

const (
	goRoutines = 100
)

func TestLockSerializaPorTenant(t *testing.T) {
	e, err := New(Config{Store: storeFalso{}})
	if err != nil {
		t.Fatal(err)
	}

	tenant := Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	var contador int
	var wg sync.WaitGroup

	for range goRoutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release := e.lock(tenant)
			defer release()
			contador++
		}()
	}

	wg.Wait()

	if contador != goRoutines {
		t.Errorf("contador = %d, want %d", contador, goRoutines)
	}

}

func TestEqualIDFactura(t *testing.T) {
	ahora := time.Now()

	testCases := []struct {
		name     string
		id1      IDFactura
		id2      IDFactura
		expected bool
	}{
		{
			name:     "Same day but different minutes",
			id1:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))},
			id2:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(time.Date(2023, 1, 1, 0, 30, 0, 0, time.FixedZone("CET", 3600)))},
			expected: true,
		},
		{
			name:     "Same time but expresed in different time zones",
			id1:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(time.Date(2023, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)))},
			id2:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(time.Date(2022, 12, 31, 23, 0, 0, 0, time.FixedZone("UTC", 0)))},
			expected: false,
		},
		{
			name:     "Time now",
			id1:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(ahora)},
			id2:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(time.Date(ahora.Year(), ahora.Month(), ahora.Day(), 0, 0, 0, 0, ahora.Location()))},
			expected: true,
		},
		{
			name:     "Different days",
			id1:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(time.Date(2023, 1, 1, 0, 0, 0, 0, time.FixedZone("CET", 3600)))},
			id2:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(time.Date(2023, 1, 2, 0, 0, 0, 0, time.FixedZone("CET", 3600)))},
			expected: false,
		},
		{
			name:     "Different NIF",
			id1:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(ahora)},
			id2:      IDFactura{NIF: "87654321B", NumSerie: "001", Fecha: record.Fecha(ahora)},
			expected: false,
		},
		{
			name:     "Different NumSerie",
			id1:      IDFactura{NIF: "12345678A", NumSerie: "001", Fecha: record.Fecha(ahora)},
			id2:      IDFactura{NIF: "12345678A", NumSerie: "002", Fecha: record.Fecha(ahora)},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.id1.Equal(tc.id2)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestLockNoBloqueaOtrosTenants(t *testing.T) {

	e, err := New(Config{Store: storeFalso{}})
	if err != nil {
		t.Fatal(err)
	}

	tenant1 := Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	var wg sync.WaitGroup

	defer wg.Wait()

	releaseA := e.lock(tenant1)

	wg.Add(1)

	defer releaseA()

	go func() {
		defer wg.Done()
		lockedTenant := e.lock(tenant1)
		lockedTenant()
	}()

	time.Sleep(50 * time.Millisecond)

	tenant2 := Tenant{NIF: "89890002K", IDSistemaInformatico: "01"}

	var hecho = make(chan struct{})

	go func() {
		releaseB := e.lock(tenant2)
		defer releaseB()
		close(hecho)
	}()

	select {
	case <-hecho:
	case <-time.After(2 * time.Second):
		t.Fatal("tenant2 was blocked by tenant1 lock, but it should not have been")
	}
}
