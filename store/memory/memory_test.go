package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/cristianemek/go-verifactu"
)

func buildEntry(secuencia uint64, numeroSerie string) *verifactu.Entry {
	return &verifactu.Entry{
		Secuencia: secuencia,
		Operacion: verifactu.OperacionAlta,
		IDFactura: verifactu.IDFactura{
			NumSerie: numeroSerie,
		},
	}
}

func TestUltimaCadenaVacia(t *testing.T) {
	s := New()

	tenant := verifactu.Tenant{
		NIF:                  "89890001K",
		IDSistemaInformatico: "01",
	}

	_, err := s.Ultimo(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Ultimo() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}
}

func TestAnexarYRecuperar(t *testing.T) {
	s := New()

	tenant := verifactu.Tenant{
		NIF:                  "89890001K",
		IDSistemaInformatico: "01",
	}

	entry := buildEntry(1, "12345678/G33")
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

	tenant := verifactu.Tenant{
		NIF:                  "89890001K",
		IDSistemaInformatico: "01",
	}

	entry := buildEntry(1, "12345678/G33")
	err := s.Anexar(context.Background(), tenant, entry)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	entryIncorrecta := buildEntry(3, "87654321/G33")
	err = s.Anexar(context.Background(), tenant, entryIncorrecta)
	if !errors.Is(err, verifactu.ErrCadenaBifurcada) {
		t.Fatalf("Anexar() = %v, want error", err)
	}
}

func TestDuplicado(t *testing.T) {
	s := New()

	tenant := verifactu.Tenant{
		NIF:                  "89890001K",
		IDSistemaInformatico: "01",
	}

	entry := buildEntry(1, "12345678/G33")

	err := s.Anexar(context.Background(), tenant, entry)

	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	entry2 := buildEntry(2, "12345678/G33")

	err = s.Anexar(context.Background(), tenant, entry2)

	if !errors.Is(err, verifactu.ErrDuplicado) {
		t.Fatalf("Expecting error = %v, but got %v", verifactu.ErrDuplicado, err)
	}

}

func TestIsolatedTenants(t *testing.T) {
	s := New()
	tenant1 := verifactu.Tenant{
		NIF:                  "89890001K",
		IDSistemaInformatico: "01",
	}

	entry1 := buildEntry(1, "12345678/G33")
	err := s.Anexar(context.Background(), tenant1, entry1)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	tenant2 := verifactu.Tenant{
		NIF:                  "89890002K",
		IDSistemaInformatico: "02",
	}

	_, err = s.Ultimo(context.Background(), tenant2)
	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Ultimo() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}

}
