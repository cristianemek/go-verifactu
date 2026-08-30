package memory

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
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

func anexarCadena(t *testing.T, s *Store, tenant verifactu.Tenant, n int) {
	t.Helper()
	for i := 1; i <= n; i++ {
		err := s.Anexar(context.Background(), tenant, buildEntry(uint64(i), fmt.Sprintf("12345678/G%d", i), verifactu.OperacionAlta))
		if err != nil {
			t.Fatalf("Anexar() = %v, want nil", err)
		}
	}
}

func buildLinea(secuencia uint64, estado record.EstadoRegistro) verifactu.LineaEnvio {
	return verifactu.LineaEnvio{
		Secuencia: secuencia,
		Estado:    estado,
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

func TestCorrecion(t *testing.T) {
	s := New()

	tenant := buildTenant("89890001K")

	entry := buildEntry(1, "12345678/G33", verifactu.OperacionAlta)

	err := s.Anexar(context.Background(), tenant, entry)

	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	entryCorregido := buildEntry(2, "12345678/G33", verifactu.OperacionAlta)
	entryCorregido.Correccion = true

	err = s.Anexar(context.Background(), tenant, entryCorregido)

	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	lastEntry, err := s.Ultimo(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Ultimo() = %v, want nil", err)
	}

	if lastEntry.Secuencia != entryCorregido.Secuencia {
		t.Errorf("Ultimo() = %v, want %v", lastEntry.Secuencia, entryCorregido.Secuencia)
	}

	if lastEntry.Correccion != true {
		t.Errorf("Ultimo() = %v, want true", lastEntry.Correccion)
	}
}

func TestUltimoEnvioVacio(t *testing.T) {
	s := New()

	tenant := buildTenant("89890001K")

	_, err := s.UltimoEnvio(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Ultimo() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}
}

func TestPendientes(t *testing.T) {
	testCases := []struct {
		name   string
		lineas []verifactu.LineaEnvio
		limite int
		want   []uint64
	}{
		{
			name:   "Sin envio",
			lineas: []verifactu.LineaEnvio{},
			limite: 0,
			want:   []uint64{1, 2, 3},
		},
		{
			name: "Todas rechazadas",
			lineas: []verifactu.LineaEnvio{
				buildLinea(1, record.EstadoRegistroIncorrecto),
				buildLinea(2, record.EstadoRegistroIncorrecto),
				buildLinea(3, record.EstadoRegistroIncorrecto),
			},
			limite: 0,
			want:   []uint64{1, 2, 3},
		},
		{
			name: "Aceptada con errores liquida",
			lineas: []verifactu.LineaEnvio{
				buildLinea(1, record.EstadoRegistroAceptadoConErrores),
				buildLinea(2, record.EstadoRegistroAceptadoConErrores),
				buildLinea(3, record.EstadoRegistroAceptadoConErrores),
			},
			limite: 0,
			want:   []uint64{},
		},
		{
			name: "Todas correctas",
			lineas: []verifactu.LineaEnvio{
				buildLinea(1, record.EstadoRegistroCorrecto),
				buildLinea(2, record.EstadoRegistroCorrecto),
				buildLinea(3, record.EstadoRegistroCorrecto),
			},
			limite: 0,
			want:   []uint64{},
		},
		{
			name: "Hueco en el envio",
			lineas: []verifactu.LineaEnvio{
				buildLinea(2, record.EstadoRegistroCorrecto),
			},
			limite: 0,
			want:   []uint64{1, 3},
		},
		{
			name:   "El limite corta la lista",
			lineas: []verifactu.LineaEnvio{},
			limite: 2,
			want:   []uint64{1, 2},
		},
		{
			name:   "El limite mayor que la cadena",
			lineas: []verifactu.LineaEnvio{},
			limite: 10,
			want:   []uint64{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := New()

			tenant := buildTenant("89890001K")
			anexarCadena(t, s, tenant, 3)

			if len(tc.lineas) > 0 {

				if len(tc.lineas) > 3 {
					t.Fatalf("Test case %s has more lines than the chain", tc.name)
				}
				err := s.AnexarEnvio(context.Background(), tenant, &verifactu.Envio{
					Lineas: tc.lineas,
				})
				if err != nil {
					t.Fatalf("AnexarEnvio() = %v, want nil", err)
				}
			}

			pendientes, err := s.Pendientes(context.Background(), tenant, tc.limite)
			if err != nil {
				t.Fatalf("Pendientes() = %v, want nil", err)
			}

			if len(pendientes) != len(tc.want) {
				t.Fatalf("Pendientes() = %v, want %v", len(pendientes), len(tc.want))
			}

			for i, secuencia := range tc.want {
				if pendientes[i].Secuencia != secuencia {
					t.Errorf("Pendientes() = %v, want %v", pendientes[i].Secuencia, secuencia)
				}
			}
		})
	}
}
