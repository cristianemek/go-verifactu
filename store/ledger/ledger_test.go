package ledger

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/store/storetest"
)

func TestFicheroValidaElTenant(t *testing.T) {
	dir := t.TempDir()

	store, err := New(dir)
	if err != nil {
		t.Fatalf("Error creating ledger: %v", err)
	}

	testCases := []struct {
		name       string
		tenant     verifactu.Tenant
		wantErr    error
		wantNombre string
	}{
		{
			name:       "tenant válido",
			tenant:     verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"},
			wantErr:    nil,
			wantNombre: "89890001K-01.jsonl",
		},
		{
			name:       "nif en minúsculas",
			tenant:     verifactu.Tenant{NIF: "89890001k", IDSistemaInformatico: "01"},
			wantErr:    verifactu.ErrTenantInvalido,
			wantNombre: "",
		},
		{
			name:       "nif vacio",
			tenant:     verifactu.Tenant{NIF: "", IDSistemaInformatico: "01"},
			wantErr:    verifactu.ErrTenantInvalido,
			wantNombre: "",
		},
		{
			name:       "idsistemainformatico vacio",
			tenant:     verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: ""},
			wantErr:    verifactu.ErrTenantInvalido,
			wantNombre: "",
		},
		{
			name:       "recorrido de rutas",
			tenant:     verifactu.Tenant{NIF: "../../etc", IDSistemaInformatico: "01"},
			wantErr:    verifactu.ErrTenantInvalido,
			wantNombre: "",
		},
		{
			name:       "caracter no ASCII",
			tenant:     verifactu.Tenant{NIF: "89890001Ñ", IDSistemaInformatico: "01"},
			wantErr:    verifactu.ErrTenantInvalido,
			wantNombre: "",
		},
		{
			name:       "NIF con guion",
			tenant:     verifactu.Tenant{NIF: "89890001-K", IDSistemaInformatico: "01"},
			wantErr:    verifactu.ErrTenantInvalido,
			wantNombre: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ruta, err := store.fichero(tc.tenant)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("fichero() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantNombre != "" && tc.wantNombre != filepath.Base(ruta) {
				t.Errorf("fichero() nombre = %v, wantNombre %v", ruta, tc.wantNombre)
				return
			}
		})
	}
}

func TestConformidad(t *testing.T) {
	storetest.Conformidad(t, func(t *testing.T) verifactu.Store {
		s, err := New(t.TempDir())
		if err != nil {
			t.Fatalf("New() = %v", err)
		}
		return s
	})
}

func TestNewDirectorio(t *testing.T) {
	escribir := func(t *testing.T, ruta, contenido string) {
		t.Helper()
		if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
			t.Fatalf("WriteFile() = %v", err)
		}
	}

	testCases := []struct {
		name     string
		preparar func(t *testing.T, dir string) string
		wantErr  bool
	}{
		{
			name: "ignora subdirectorios, otros ficheros y lineas vacias",
			preparar: func(t *testing.T, dir string) string {
				if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
					t.Fatalf("Mkdir() = %v", err)
				}
				escribir(t, filepath.Join(dir, "notas.txt"), "hola")
				escribir(t, filepath.Join(dir, "89890001K-01.jsonl"), "\n")
				return dir
			},
		},
		{
			name: "nombre que no es NIF-ID",
			preparar: func(t *testing.T, dir string) string {
				escribir(t, filepath.Join(dir, "malo.jsonl"), "")
				return dir
			},
			wantErr: true,
		},
		{
			name: "cadena corrupta",
			preparar: func(t *testing.T, dir string) string {
				escribir(t, filepath.Join(dir, "89890001K-01.jsonl"), "no es json\n")
				return dir
			},
			wantErr: true,
		},
		{
			name: "envios corruptos",
			preparar: func(t *testing.T, dir string) string {
				escribir(t, filepath.Join(dir, "89890001K-01.envios.jsonl"), "no es json\n")
				return dir
			},
			wantErr: true,
		},
		{
			name: "la ruta es un fichero",
			preparar: func(t *testing.T, dir string) string {
				ruta := filepath.Join(dir, "fichero")
				escribir(t, ruta, "")
				return ruta
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.preparar(t, t.TempDir()))

			if tc.wantErr && err == nil {
				t.Fatal("New() = nil, want error")
			}

			if !tc.wantErr && err != nil {
				t.Fatalf("New() = %v, want nil", err)
			}
		})
	}
}

func TestAnexarNoPuedeEscribir(t *testing.T) {
	ctx := context.Background()
	valido := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}
	invalido := verifactu.Tenant{NIF: "../x", IDSistemaInformatico: "01"}

	dir := t.TempDir()

	s, err := New(dir)
	if err != nil {
		t.Fatalf("New() = %v", err)
	}

	for _, nombre := range []string{"89890001K-01.jsonl", "89890001K-01.envios.jsonl"} {
		if err := os.Mkdir(filepath.Join(dir, nombre), 0o755); err != nil {
			t.Fatalf("Mkdir() = %v", err)
		}
	}

	entry := &verifactu.Entry{Secuencia: 1, Operacion: verifactu.OperacionAlta}

	if err := s.Anexar(ctx, invalido, entry); !errors.Is(err, verifactu.ErrTenantInvalido) {
		t.Errorf("Anexar(tenant invalido) = %v, want ErrTenantInvalido", err)
	}

	if err := s.AnexarEnvio(ctx, invalido, &verifactu.Envio{}); !errors.Is(err, verifactu.ErrTenantInvalido) {
		t.Errorf("AnexarEnvio(tenant invalido) = %v, want ErrTenantInvalido", err)
	}

	if err := s.Anexar(ctx, valido, entry); err == nil {
		t.Error("Anexar() sobre un directorio = nil, want error")
	}

	if err := s.AnexarEnvio(ctx, valido, &verifactu.Envio{}); err == nil {
		t.Error("AnexarEnvio() sobre un directorio = nil, want error")
	}
}
