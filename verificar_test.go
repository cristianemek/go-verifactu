package verifactu_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

	_, err = engine.Anular(context.Background(), tenant, validRegistroAnulacion("001"))

	if err != nil {
		t.Fatal(err)
	}

	entry, err := store.Cadena(context.Background(), tenant)

	if err != nil {
		t.Fatal(err)
	}

	return entry
}

func TestVerificarCadenaValida(t *testing.T) {
	c := cadenaDePrueba(t)

	if len(c) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(c))
	}

	if err := verifactu.VerificarCadena(c); err != nil {
		t.Fatalf("expected valid chain, got error: %v", err)
	}

	if err := verifactu.VerificarCadena(nil); err != nil {
		t.Fatalf("expected valid chain for nil, got error: %v", err)
	}
}

func rehacerHuella(e *verifactu.Entry) {
	if e.Alta != nil {
		e.Huella = e.Alta.Fingerprint()
		e.Alta.Huella = e.Huella
	}

	if e.Anulacion != nil {
		e.Huella = e.Anulacion.Fingerprint()
		e.Anulacion.Huella = e.Huella
	}
}

func TestVerificarCadenaRota(t *testing.T) {
	testCases := []struct {
		name    string
		breakFn func([]*verifactu.Entry) []*verifactu.Entry
		wantMsg string
	}{
		{
			name: "secuencia alterada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Secuencia = 5
				return entries
			},
			wantMsg: "secuencia 5 in position 1",
		},
		{
			name: "huella alterada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Huella = "altered"
				return entries
			},
			wantMsg: "mismatched fingerprint for",
		},
		{
			name: "registro alterado",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Alta.CuotaTotal = record.Amount(100)
				return entries
			},
			wantMsg: "mismatched fingerprint for",
		},
		{
			name: "entrada eliminada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				return []*verifactu.Entry{entries[0], entries[2]}
			},
			wantMsg: "secuencia 3 in position 1",
		},
		{
			name: "huella de Alta alterada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Alta.Huella = "altered"
				return entries
			},
			wantMsg: "mismatched fingerprint in entry",
		},
		{
			name: "huella de Anulacion alterada",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[3].Anulacion.Huella = "altered"
				return entries
			},
			wantMsg: "mismatched fingerprint in entry",
		},
		{
			name: "Alta a nil",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Alta = nil
				return entries
			},
			wantMsg: "missing Alta",
		},
		{
			name: "Anulacion a nil",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[3].Anulacion = nil
				return entries
			},
			wantMsg: "missing Anulacion",
		},
		{
			name: "operacion desconocida",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Operacion = "otra"
				return entries
			},
			wantMsg: "unknown operation",
		},
		{
			name: "primera sin PrimerRegistro",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[0].Alta.Encadenamiento.PrimerRegistro = nil
				rehacerHuella(entries[0])
				return entries
			},
			wantMsg: "non-nil PrimerRegistro",
		},
		{
			name: "primera con RegistroAnterior",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				anterior := *entries[1].Alta.Encadenamiento.RegistroAnterior
				entries[0].Alta.Encadenamiento.RegistroAnterior = &anterior
				rehacerHuella(entries[0])
				return entries
			},
			wantMsg: "must have a nil RegistroAnterior",
		},
		{
			name: "sin RegistroAnterior",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				entries[1].Alta.Encadenamiento.RegistroAnterior = nil
				rehacerHuella(entries[1])
				return entries
			},
			wantMsg: "must have a non-nil RegistroAnterior",
		},
		{
			name: "RegistroAnterior apunta a otra huella",
			breakFn: func(entries []*verifactu.Entry) []*verifactu.Entry {
				anterior := *entries[1].Alta.Encadenamiento.RegistroAnterior
				anterior.Huella = strings.Repeat("A", 64)
				entries[1].Alta.Encadenamiento.RegistroAnterior = &anterior
				rehacerHuella(entries[1])
				return entries
			},
			wantMsg: "mismatched RegistroAnterior",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := cadenaDePrueba(t)

			err := verifactu.VerificarCadena(tc.breakFn(c))

			if !errors.Is(err, verifactu.ErrCadenaBifurcada) {
				t.Fatalf("expected error %v, got %v", verifactu.ErrCadenaBifurcada, err)
			}

			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantMsg)
			}
		})
	}
}
