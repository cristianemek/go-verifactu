package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/cristianemek/go-verifactu"
)

func buildEntry(secuencia uint64, numeroSerie string, operacion verifactu.Operacion) *verifactu.Entry {
	return &verifactu.Entry{
		Secuencia: secuencia,
		Operacion: operacion,
		IDFactura: verifactu.IDFactura{
			NumSerie: numeroSerie,
		},
	}
}

func buildTenant(nif string) verifactu.Tenant {

	return verifactu.Tenant{
		NIF:                  nif,
		IDSistemaInformatico: "01",
	}
}

func TestUltimaCadenaVacia(t *testing.T) {
	s := New()

	tenant := buildTenant("89890001K")

	_, err := s.Ultimo(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Ultimo() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}
}

func TestAnexarYRecuperar(t *testing.T) {
	s := New()

	tenant := buildTenant("89890001K")

	entry := buildEntry(1, "12345678/G33", verifactu.OperacionAlta)
	err := s.Anexar(context.Background(), tenant, entry)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	recuperado, err := s.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Ultimo() = %v, want nil", err)
	}

	if recuperado.Secuencia != entry.Secuencia {
		t.Errorf("Ultimo() = %v, want %v", recuperado.Secuencia, entry.Secuencia)
	}

	if recuperado.IDFactura.NumSerie != entry.IDFactura.NumSerie {
		t.Errorf("Ultimo() = %v, want %v", recuperado.IDFactura.NumSerie, entry.IDFactura.NumSerie)
	}
}

func TestSecuenciasIncorrectas(t *testing.T) {
	s := New()

	tenant := buildTenant("89890001K")

	entry := buildEntry(1, "12345678/G33", verifactu.OperacionAlta)
	err := s.Anexar(context.Background(), tenant, entry)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	entryIncorrecta := buildEntry(3, "87654321/G33", verifactu.OperacionAlta)
	err = s.Anexar(context.Background(), tenant, entryIncorrecta)
	if !errors.Is(err, verifactu.ErrCadenaBifurcada) {
		t.Fatalf("Anexar() = %v, want error", err)
	}
}

func TestDuplicado(t *testing.T) {
	s := New()

	tenant := buildTenant("89890001K")

	entry := buildEntry(1, "12345678/G33", verifactu.OperacionAlta)

	err := s.Anexar(context.Background(), tenant, entry)

	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	entry2 := buildEntry(2, "12345678/G33", verifactu.OperacionAlta)

	err = s.Anexar(context.Background(), tenant, entry2)

	if !errors.Is(err, verifactu.ErrDuplicado) {
		t.Fatalf("Expecting error = %v, but got %v", verifactu.ErrDuplicado, err)
	}

}

func TestIsolatedTenants(t *testing.T) {
	s := New()
	tenant1 := buildTenant("89890001K")

	entry1 := buildEntry(1, "12345678/G33", verifactu.OperacionAlta)
	err := s.Anexar(context.Background(), tenant1, entry1)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	tenant2 := buildTenant("89890002K")
	tenant2.IDSistemaInformatico = "02"

	_, err = s.Ultimo(context.Background(), tenant2)
	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Ultimo() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}

}

func TestBusquedas(t *testing.T) {

	s := New()

	tenant := buildTenant("89890001K")

	entry1 := buildEntry(1, "12345678/G33", verifactu.OperacionAlta)
	entry2 := buildEntry(2, "12345678/G33", verifactu.OperacionAnulacion)
	entry3 := buildEntry(3, "87654321/G33", verifactu.OperacionAlta)

	err := s.Anexar(context.Background(), tenant, entry1)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	err = s.Anexar(context.Background(), tenant, entry2)

	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	err = s.Anexar(context.Background(), tenant, entry3)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	testCases := []struct {
		name    string
		id      verifactu.IDFactura
		op      verifactu.Operacion
		want    *verifactu.Entry
		wantErr error
	}{
		{
			name:    "Buscar entry1",
			id:      entry1.IDFactura,
			op:      verifactu.OperacionAlta,
			want:    entry1,
			wantErr: nil,
		},
		{
			name:    "Buscar entry1 anulado",
			id:      entry1.IDFactura,
			op:      verifactu.OperacionAnulacion,
			want:    entry2,
			wantErr: nil,
		},
		{
			name:    "Buscar entry3",
			id:      entry3.IDFactura,
			op:      verifactu.OperacionAlta,
			want:    entry3,
			wantErr: nil,
		},
		{
			name: "Buscar NumSerie inexistente",
			id: verifactu.IDFactura{
				NumSerie: "00000000/G33",
			},
			op:      verifactu.OperacionAlta,
			want:    nil,
			wantErr: verifactu.ErrNoEncontrado,
		},
		{
			name:    "Buscar Operacion inexistente",
			id:      entry3.IDFactura,
			op:      verifactu.OperacionAnulacion,
			want:    nil,
			wantErr: verifactu.ErrNoEncontrado,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := s.Buscar(context.Background(), tenant, tc.id, tc.op)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Buscar() = %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("Buscar() = %v, want %v", got, tc.want)
			}
		})
	}

	s2 := New()

	got, err := s2.Buscar(context.Background(), tenant, entry1.IDFactura, entry1.Operacion)

	if err == nil {
		t.Fatalf("Buscar() = %v, want error", err)
	}

	if got != nil {
		t.Errorf("Buscar() = %v, want nil", got)
	}

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Buscar() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}
}
