package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func escribirConfig(t *testing.T, nif, token string, store string) string {
	dir := t.TempDir()
	config := fmt.Sprintf(`{
			"listen": ":8080",
			"data": "./datos",
			"entorno": "pruebas",
			"remision_cada": "60s",
			"store": "%s",
			"sistema": {
				"NombreRazon": "TU EMPRESA SL",
				"NIF": "89890001K",
				"NombreSistemaInformatico": "verifactud",
				"IdSistemaInformatico": "01",
				"Version": "0.9.0",
				"NumeroInstalacion": "1",
				"TipoUsoPosibleSoloVerifactu": "S",
				"TipoUsoPosibleMultiOT": "S",
				"IndicadorMultiplesOT": "S"
			},
			"tenants": {
				"%s": {
				"nombre": "EMPRESA DE PRUEBAS SL",
				"token": "%s",
				"certificado": "dev.pem",
				"tipo_certificado": "representante"
				}
			}
			}`, store, nif, token)

	path := filepath.Join(dir, "config.json")

	err := os.WriteFile(path, []byte(config), 0644)
	if err != nil {
		t.Fatalf("Error writing config file: %v", err)
	}
	return path
}

func TestCargarConfigRechazaTokenVacio(t *testing.T) {

	testCases := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "con token",
			token:   "token123",
			wantErr: false,
		},
		{
			name:    "sin token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := escribirConfig(t, "89890001K", tc.token, "ledger")

			_, err := cargarConfig(path)

			if (tc.wantErr && err == nil) || (!tc.wantErr && err != nil) {
				t.Errorf("cargarConfig() error = %v, wantErr %v", err, tc.wantErr)
			}

		})

	}
}

func TestCargarConfigNIFEnMayusculas(t *testing.T) {
	path := escribirConfig(t, "89890001k", "token123", "ledger")

	cfg, err := cargarConfig(path)

	if err != nil {
		t.Fatalf("cargarConfig() error = %v", err)
	}

	if _, ok := cfg.Tenants["89890001K"]; !ok {
		t.Errorf("cargarConfig() did not convert NIF to uppercase, expected key '89890001K'")
	}

	if cfg.Log != "" {
		t.Errorf("cargarConfig() log path = %s, want empty string", cfg.Log)
	}
}

func TestCargarConfigExigeCertificadoPorTenant(t *testing.T) {
	cfg := `{
			"listen": ":8080",
			"data": "./datos",
			"entorno": "pruebas",
			"remision_cada": "60s",
			"sistema": {
				"NombreRazon": "TU EMPRESA SL",
				"NIF": "89890001K",
				"NombreSistemaInformatico": "verifactud",
				"IdSistemaInformatico": "01",
				"Version": "0.9.0",
				"NumeroInstalacion": "1",
				"TipoUsoPosibleSoloVerifactu": "S",
				"TipoUsoPosibleMultiOT": "S",
				"IndicadorMultiplesOT": "S"
			},
			"tenants": {
				"89890001K": {
				"nombre": "EMPRESA DE PRUEBAS SL",
				"token": "token123",
				"certificado": "dev.pem",
				"tipo_certificado": "representante"
				},
				"89890002F": {
				"nombre": "EMPRESA DE PRUEBAS2 SL",
				"token": "token123"
				}
			}
			}`

	path := filepath.Join(t.TempDir(), "config.json")

	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		t.Fatalf("Error writing config file: %v", err)
	}

	_, err := cargarConfig(path)

	if err == nil || !strings.Contains(err.Error(), "89890002F") {
		t.Fatalf("cargarConfig() error = %v, want one naming 89890002F", err)
	}

}

func TestStoreType(t *testing.T) {
	testCases := []struct {
		name    string
		store   string
		wantErr bool
	}{
		{
			name:    "valid ledger store",
			store:   "ledger",
			wantErr: false,
		},
		{
			name:    "un supported store type",
			store:   "postgres",
			wantErr: true,
		},
		{
			name:    "valid sqlite store",
			store:   "sqlite",
			wantErr: false,
		},
		{
			name:    "empty store type defaults to ledger",
			store:   "",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := escribirConfig(t, "89890001K", "token123", tc.store)

			cfg, err := cargarConfig(path)

			if (tc.wantErr && err == nil) || (!tc.wantErr && err != nil) {
				t.Errorf("cargarConfig() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.store == "" && err == nil {
				if cfg.Store != "ledger" {
					t.Errorf("cargarConfig() store = %s, want default 'ledger'", cfg.Store)
				}

			}

		})

	}

}
