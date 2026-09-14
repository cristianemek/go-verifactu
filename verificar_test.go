package verifactu_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/memory"
)

func cadenaDePrueba(t *testing.T) []*verifactu.Entry {

	t.Helper()

	store := memory.New()

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})

	if err != nil {
		t.Fatal(err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	for i := range 3 {
		_, err := engine.Alta(context.Background(), tenant, validRegistroAlta(fmt.Sprintf("%03d", i+1)))
		if err != nil {
			t.Fatal(err)
		}
	}

	entry, err := store.Cadena(context.Background(), tenant)

	if err != nil {
		t.Fatal(err)
	}

	return entry
}

func TestVerificarCadenaValida(t *testing.T) {
	c := cadenaDePrueba(t)

	if len(c) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(c))
	}

	if err := verifactu.VerificarCadena(c); err != nil {
		t.Fatalf("expected valid chain, got error: %v", err)
	}

	if err := verifactu.VerificarCadena(nil); err != nil {
		t.Fatalf("expected valid chain for nil, got error: %v", err)
	}
}

func TestVerificarCadenaRota(t *testing.T) {
	testCases := []struct {
		name    string
		breakFn func([]*verifactu.Entry) []*verifactu.Entry
	}{
		{
			name: "secuencia alterada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Secuencia = 5
				return entries
			},
		},
		{
			name: "huella alterada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Huella = "altered"
				return entries
			},
		},
		{
			name: "registro alterado",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Alta.CuotaTotal = record.Amount(100)
				return entries
			},
		},
		{
			name: "entrada eliminada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				return []*verifactu.Entry{entries[0], entries[2]}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := cadenaDePrueba(t)

			err := verifactu.VerificarCadena(tc.breakFn(c))

			if !errors.Is(err, verifactu.ErrCadenaBifurcada) {
				t.Fatalf("expected error %v, got %v", verifactu.ErrCadenaBifurcada, err)
			}
		})
	}

}
