package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

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
			dir := t.TempDir()
			config := fmt.Sprintf(`{
			"listen": ":8080",
			"data": "./datos",
			"entorno": "pruebas",
			"certificado": "dev.pem",
			"tipo_certificado": "representante",
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
				"token": "%s"
				}
			}
			}`, tc.token)

			path := filepath.Join(dir, "config.json")

			err := os.WriteFile(path, []byte(config), 0644)
			if err != nil {
				t.Fatalf("Error writing config file: %v", err)
			}

			_, err = cargarConfig(path)

			if (tc.wantErr && err == nil) || (!tc.wantErr && err != nil) {
				t.Errorf("cargarConfig() error = %v, wantErr %v", err, tc.wantErr)
			}

		})

	}
}
