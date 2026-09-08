package verifactu_test

import (
	"context"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

type storeConflictivo struct {
	llamadasUltimo int
	anexados       []*verifactu.Entry
	anterior       *verifactu.Entry
}

// Anexar implements [verifactu.Store].
func (s *storeConflictivo) Anexar(ctx context.Context, t verifactu.Tenant, e *verifactu.Entry) error {
	s.anexados = append(s.anexados, e)

	if len(s.anexados) == 1 {
		return verifactu.ErrConflictoDeSecuencia
	}

	return nil
}

// AnexarEnvio implements [verifactu.Store].
func (s *storeConflictivo) AnexarEnvio(ctx context.Context, t verifactu.Tenant, envio *verifactu.Envio) error {
	panic("unimplemented")
}

// Buscar implements [verifactu.Store].
func (s *storeConflictivo) Buscar(ctx context.Context, t verifactu.Tenant, idFactura verifactu.IDFactura, op verifactu.Operacion) (*verifactu.Entry, error) {
	return nil, verifactu.ErrNoEncontrado
}

// Pendientes implements [verifactu.Store].
func (s *storeConflictivo) Pendientes(ctx context.Context, t verifactu.Tenant, limite int) ([]*verifactu.Entry, error) {
	panic("unimplemented")
}

// Ultimo implements [verifactu.Store].
func (s *storeConflictivo) Ultimo(ctx context.Context, t verifactu.Tenant) (*verifactu.Entry, error) {
	s.llamadasUltimo++

	if s.llamadasUltimo == 1 {
		return nil, verifactu.ErrNoEncontrado
	}
	return s.anterior, nil
}

// UltimoEnvio implements [verifactu.Store].
func (s *storeConflictivo) UltimoEnvio(ctx context.Context, t verifactu.Tenant) (*verifactu.Envio, error) {
	panic("unimplemented")
}

var _ verifactu.Store = (*storeConflictivo)(nil)

func TestAltaReintentaTrasConflicto(t *testing.T) {
	anterior := &verifactu.Entry{
		Secuencia: 1,
		Huella:    "huella",
		IDFactura: verifactu.IDFactura{
			NIF:      "12345678A",
			NumSerie: "1",
			Fecha:    record.Fecha(time.Now())},
	}

	store := &storeConflictivo{
		anterior: anterior,
	}

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatal(err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	_, err = engine.Alta(context.Background(), tenant, validRegistroAlta("001"))

	if err != nil {
		t.Fatalf("Alta() = %v, want nil", err)
	}

	if len(store.anexados) != 2 {
		t.Fatalf("Anexar() called %d times, want 2", len(store.anexados))
	}

	if store.anexados[0].Secuencia != 1 {
		t.Errorf("Anexar() first call Secuencia = %d, want 1", store.anexados[0].Secuencia)
	}

	if store.anexados[1].Secuencia != 2 {
		t.Errorf("Anexar() second call Secuencia = %d, want 2", store.anexados[1].Secuencia)
	}

	if store.anexados[0].Huella == store.anexados[1].Huella {
		t.Errorf("Anexar() first and second call Huella = %s, want different", store.anexados[0].Huella)
	}

	if store.anexados[1].Alta.Encadenamiento.RegistroAnterior.Huella != anterior.Huella {
		t.Errorf("Anexar() second call RegistroAnterior.Huella = %s, want %s", store.anexados[1].Alta.Encadenamiento.RegistroAnterior.Huella, anterior.Huella)
	}
}
