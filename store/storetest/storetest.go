package storetest

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

// Conformidad runs the shared suite against the Store built by nuevo.
// nuevo is called once per subtest, so each one starts empty.
func Conformidad(t *testing.T, nuevo func(t *testing.T) verifactu.Store) {
	t.Run("Ultimo con la cadena vacia", func(t *testing.T) {
		testUltimaCadenaVacia(t, nuevo(t))
	})
	t.Run("UltimoEnvio sin envios", func(t *testing.T) {
		testUltimoEnvioVacio(t, nuevo(t))
	})
	t.Run("Anexar y recuperar", func(t *testing.T) {
		testAnexarYRecuperar(t, nuevo(t))
	})
	t.Run("Conflicto de secuencia", func(t *testing.T) {
		testConflictoDeSecuencia(t, nuevo(t))
	})
	t.Run("Duplicado", func(t *testing.T) {
		testDuplicado(t, nuevo(t))
	})
	t.Run("Tenants aislados", func(t *testing.T) {
		testIsolatedTenants(t, nuevo(t))
	})
	t.Run("Busquedas", func(t *testing.T) {
		testBusquedas(t, nuevo(t))
	})
	t.Run("Correccion", func(t *testing.T) {
		testCorrecion(t, nuevo(t))
	})
	t.Run("Pendientes", func(t *testing.T) {
		testPendientes(t, nuevo)
	})
	t.Run("EnvioDe", func(t *testing.T) {
		testEnvioDe(t, nuevo(t))
	})
}

func buildEntry(secuencia uint64, numeroSerie string, operacion verifactu.Operacion) *verifactu.Entry {
	return &verifactu.Entry{
		Secuencia: secuencia,
		Operacion: operacion,
		IDFactura: verifactu.IDFactura{
			NumSerie: numeroSerie,
		},
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

func anexarCadena(t *testing.T, s verifactu.Store, tenant verifactu.Tenant, n int) {
	t.Helper()
	for i := 1; i <= n; i++ {
		err := s.Anexar(context.Background(), tenant, buildEntry(uint64(i), fmt.Sprintf("12345678/G%d", i), verifactu.OperacionAlta))
		if err != nil {
			t.Fatalf("Anexar() = %v, want nil", err)
		}
	}
}

func testUltimaCadenaVacia(t *testing.T, s verifactu.Store) {
	tenant := buildTenant("89890001K")

	_, err := s.Ultimo(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Ultimo() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}
}

func testUltimoEnvioVacio(t *testing.T, s verifactu.Store) {

	tenant := buildTenant("89890001K")

	_, err := s.UltimoEnvio(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("UltimoEnvio() = %v, want %v", err, verifactu.ErrNoEncontrado)
	}
}

func testAnexarYRecuperar(t *testing.T, s verifactu.Store) {

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

func testConflictoDeSecuencia(t *testing.T, s verifactu.Store) {
	tenant := buildTenant("89890001K")

	entry := buildEntry(1, "12345678/G33", verifactu.OperacionAlta)
	err := s.Anexar(context.Background(), tenant, entry)
	if err != nil {
		t.Fatalf("Anexar() = %v, want nil", err)
	}

	entryIncorrecta := buildEntry(3, "87654321/G33", verifactu.OperacionAlta)
	err = s.Anexar(context.Background(), tenant, entryIncorrecta)
	if !errors.Is(err, verifactu.ErrConflictoDeSecuencia) {
		t.Fatalf("Anexar() = %v, want error", err)
	}
}

func testDuplicado(t *testing.T, s verifactu.Store) {
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

func testIsolatedTenants(t *testing.T, s verifactu.Store) {
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

func testBusquedas(t *testing.T, s verifactu.Store) {
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

			if tc.want == nil {
				if got != nil {
					t.Errorf("Buscar() = %v, want nil", got)
				}
				return
			}

			// Se compara la secuencia y no el puntero: un store que lee de disco
			// devuelve una copia, no el mismo objeto que se anexó.
			if got == nil || got.Secuencia != tc.want.Secuencia {
				t.Errorf("Buscar() = %v, want secuencia %d", got, tc.want.Secuencia)
			}
		})
	}

	otroTenant := buildTenant("89890002K")

	got, err := s.Buscar(context.Background(), otroTenant, entry1.IDFactura, entry1.Operacion)

	if !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Fatalf("Buscar() en otro tenant = %v, want %v", err, verifactu.ErrNoEncontrado)
	}

	if got != nil {
		t.Errorf("Buscar() en otro tenant = %v, want nil", got)
	}
}

func testCorrecion(t *testing.T, s verifactu.Store) {
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

// testPendientes recibe el constructor y no un store: cada fila empieza vacía.
func testPendientes(t *testing.T, nuevo func(t *testing.T) verifactu.Store) {
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
			name: "Todas rechazadas se procesan",
			lineas: []verifactu.LineaEnvio{
				buildLinea(1, record.EstadoRegistroIncorrecto),
				buildLinea(2, record.EstadoRegistroIncorrecto),
				buildLinea(3, record.EstadoRegistroIncorrecto),
			},
			limite: 0,
			want:   []uint64{},
		},
		{
			name: "Aceptada con errores se procesa",
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
			s := nuevo(t)

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

func testEnvioDe(t *testing.T, s verifactu.Store) {
	tenant := buildTenant("89890001K")

	envio := &verifactu.Envio{
		Lineas: []verifactu.LineaEnvio{
			buildLinea(1, record.EstadoRegistroCorrecto),
			buildLinea(2, record.EstadoRegistroAceptadoConErrores),
		},
		CSV: "CSV-1",
	}

	err := s.AnexarEnvio(context.Background(), tenant, envio)
	if err != nil {
		t.Fatalf("AnexarEnvio() = %v, want nil", err)
	}

	err = s.AnexarEnvio(context.Background(), tenant, &verifactu.Envio{
		Lineas: []verifactu.LineaEnvio{
			buildLinea(3, record.EstadoRegistroIncorrecto),
		},
		CSV: "CSV-2",
	})

	if err != nil {
		t.Fatalf("AnexarEnvio() = %v, want nil", err)
	}

	testCases := []struct {
		secuencia uint64
		csv       string
		err       error
	}{
		{secuencia: 1, csv: "CSV-1"},
		{secuencia: 2, csv: "CSV-1"},
		{secuencia: 3, csv: "CSV-2"},
		{secuencia: 4, err: verifactu.ErrNoEncontrado},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Secuencia %d", tc.secuencia), func(t *testing.T) {
			envio, err := s.EnvioDe(context.Background(), tenant, tc.secuencia)

			if !errors.Is(err, tc.err) {
				t.Fatalf("EnvioDe() = %v, want %v", err, tc.err)
			}

			if tc.err != nil {
				return
			}

			if envio.CSV != tc.csv {
				t.Errorf("EnvioDe() CSV = %q, want %q", envio.CSV, tc.csv)
			}
		})
	}
}
