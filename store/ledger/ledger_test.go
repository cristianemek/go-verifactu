package ledger

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/cristianemek/go-verifactu"
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
