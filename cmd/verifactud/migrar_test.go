package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/ledger"
	"github.com/cristianemek/go-verifactu/store/sqlite"
)

func TestComandoMigrar(t *testing.T) {
	dir := t.TempDir()

	rutaLedger := filepath.Join(dir, "datos")
	rutaDB := filepath.Join(dir, "verifactu.db")

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	config := fmt.Sprintf(`{
			"listen": ":8080",
			"data": %q,
			"entorno": "pruebas",
			"remision_cada": "60s",
			"store": "ledger",
			"sistema": {
				"NombreRazon": "TU EMPRESA SL",
				"NIF": "89890001K",
				"NombreSistemaInformatico": "verifactud",
				"IdSistemaInformatico": "%s",
				"Version": "0.9.0",
				"NumeroInstalacion": "1",
				"TipoUsoPosibleSoloVerifactu": "S",
				"TipoUsoPosibleMultiOT": "S",
				"IndicadorMultiplesOT": "S"
			},
			"tenants": {
				"%s": {
				"nombre": "EMPRESA DE PRUEBAS SL",
				"token": "token123",
				"certificado": "dev.pem",
				"tipo_certificado": "representante"
				}
			}
			}`, rutaLedger, tenant.IDSistemaInformatico, tenant.NIF)

	origen, err := ledger.New(rutaLedger)
	if err != nil {
		t.Fatalf("ledger.New() = %v", err)
	}

	sistema := record.SistemaInformatico{
		NIF:                         record.Ptr("89890001K"),
		NombreRazon:                 "Sistema de Prueba",
		NombreSistemaInformatico:    "verifactu",
		IdSistemaInformatico:        "01",
		Version:                     "1.0",
		NumeroInstalacion:           "1",
		TipoUsoPosibleSoloVerifactu: record.SiNoSi,
		TipoUsoPosibleMultiOT:       record.SiNoSi,
		IndicadorMultiplesOT:        record.SiNoSi,
	}

	engine, err := verifactu.New(verifactu.Config{
		Store:              origen,
		Now:                func() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) },
		SistemaInformatico: &sistema,
	})
	if err != nil {
		t.Fatalf("verifactu.New() = %v", err)
	}

	ctx := context.Background()

	var factura record.RegistroAlta
	if err := json.Unmarshal([]byte(facturaJSON), &factura); err != nil {
		t.Fatalf("Unmarshal(facturaJSON) = %v", err)
	}

	entrada1, err := engine.Alta(ctx, tenant, factura)
	if err != nil {
		t.Fatalf("Alta(1) = %v", err)
	}

	factura.IDFactura.NumSerieFactura = "F-2026-0002"

	entrada2, err := engine.Alta(ctx, tenant, factura)
	if err != nil {
		t.Fatalf("Alta(2) = %v", err)
	}

	envio := &verifactu.Envio{
		Lineas: []verifactu.LineaEnvio{
			{
				Estado:    record.EstadoRegistroCorrecto,
				Secuencia: entrada1.Secuencia,
				Operacion: verifactu.OperacionAlta,
				IDFactura: entrada1.IDFactura,
			},
		},
		CSV:          "A-1",
		Instante:     time.Date(2026, 9, 10, 12, 0, 0, 120000001, time.UTC),
		TiempoEspera: time.Second * 60,
		EstadoEnvio:  record.EstadoEnvioCorrecto,
	}

	if err := origen.AnexarEnvio(ctx, tenant, envio); err != nil {
		t.Fatalf("AnexarEnvio() = %v", err)
	}

	rutaConfig := filepath.Join(dir, "config.json")

	if err := os.WriteFile(rutaConfig, []byte(config), 0o600); err != nil {
		t.Fatalf("os.WriteFile() = %v", err)
	}

	if err := comandoMigrar([]string{"-config", rutaConfig, "-a", rutaDB}); err != nil {
		t.Fatalf("comandoMigrar() = %v", err)
	}

	destino, err := sqlite.New(rutaDB)
	if err != nil {
		t.Fatalf("sqlite.New() = %v", err)
	}

	t.Cleanup(func() { destino.Close() })

	cadena, err := destino.Cadena(ctx, tenant)
	if err != nil {
		t.Fatalf("Cadena() = %v", err)
	}

	if len(cadena) != 2 {
		t.Fatalf("Cadena() = %d entradas, want 2", len(cadena))
	}

	if cadena[0].Huella != entrada1.Huella || cadena[1].Huella != entrada2.Huella {
		t.Errorf("huellas = %q y %q, want %q y %q", cadena[0].Huella, cadena[1].Huella, entrada1.Huella, entrada2.Huella)
	}

	pendientes, err := destino.Pendientes(ctx, tenant, 0)
	if err != nil {
		t.Fatalf("Pendientes() = %v", err)
	}

	if len(pendientes) != 1 {
		t.Fatalf("Pendientes() = %d, want 1", len(pendientes))
	}

	if pendientes[0].Secuencia != entrada2.Secuencia {
		t.Errorf("Pendientes()[0] = secuencia %d, want %d", pendientes[0].Secuencia, entrada2.Secuencia)
	}

	ultimoEnvio, err := destino.UltimoEnvio(ctx, tenant)
	if err != nil {
		t.Fatalf("UltimoEnvio() = %v", err)
	}

	if ultimoEnvio.CSV != envio.CSV {
		t.Errorf("UltimoEnvio() CSV = %q, want %q", ultimoEnvio.CSV, envio.CSV)
	}

	if !ultimoEnvio.Instante.Equal(envio.Instante) {
		t.Errorf("UltimoEnvio() Instante = %v, want %v", ultimoEnvio.Instante, envio.Instante)
	}

	if len(ultimoEnvio.Lineas) != 1 {
		t.Errorf("UltimoEnvio() = %d lineas, want 1", len(ultimoEnvio.Lineas))
	}

	if err := comandoMigrar([]string{"-config", rutaConfig, "-a", rutaDB}); err == nil {
		t.Fatal("comandoMigrar() sobre un destino existente = nil, want error")
	}
}
