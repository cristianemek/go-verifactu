package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/ledger"
	"github.com/cristianemek/go-verifactu/store/sqlite"
)

type migracion struct {
	rutaConfig  string
	rutaDB      string
	rutaLedger  string
	tenant      verifactu.Tenant
	sinFacturas verifactu.Tenant
	entrada1    *verifactu.Entry
	entrada2    *verifactu.Entry
	envio       *verifactu.Envio
}

func prepararMigracion(t *testing.T) migracion {
	t.Helper()

	dir := t.TempDir()

	m := migracion{
		rutaConfig:  filepath.Join(dir, "config.json"),
		rutaDB:      filepath.Join(dir, "verifactu.db"),
		rutaLedger:  filepath.Join(dir, "datos"),
		tenant:      verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"},
		sinFacturas: verifactu.Tenant{NIF: "89890002L", IDSistemaInformatico: "01"},
	}

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
				},
				"%s": {
				"nombre": "EMPRESA DE PRUEBAS SL",
				"token": "token456",
				"certificado": "dev.pem",
				"tipo_certificado": "representante"
				}
			}
			}`, m.rutaLedger, m.tenant.IDSistemaInformatico, m.tenant.NIF, m.sinFacturas.NIF)

	origen, err := ledger.New(m.rutaLedger)
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

	m.entrada1, err = engine.Alta(ctx, m.tenant, factura)
	if err != nil {
		t.Fatalf("Alta(1) = %v", err)
	}

	factura.IDFactura.NumSerieFactura = "F-2026-0002"

	m.entrada2, err = engine.Alta(ctx, m.tenant, factura)
	if err != nil {
		t.Fatalf("Alta(2) = %v", err)
	}

	m.envio = &verifactu.Envio{
		Lineas: []verifactu.LineaEnvio{
			{
				Estado:    record.EstadoRegistroCorrecto,
				Secuencia: m.entrada1.Secuencia,
				Operacion: verifactu.OperacionAlta,
				IDFactura: m.entrada1.IDFactura,
			},
		},
		CSV:          "A-1",
		Instante:     time.Date(2026, 9, 10, 12, 0, 0, 120000001, time.UTC),
		TiempoEspera: time.Second * 60,
		EstadoEnvio:  record.EstadoEnvioCorrecto,
	}

	if err := origen.AnexarEnvio(ctx, m.tenant, m.envio); err != nil {
		t.Fatalf("AnexarEnvio() = %v", err)
	}

	if err := os.WriteFile(m.rutaConfig, []byte(config), 0o600); err != nil {
		t.Fatalf("os.WriteFile() = %v", err)
	}

	return m
}

func TestComandoMigrar(t *testing.T) {
	m := prepararMigracion(t)
	ctx := context.Background()

	if err := comandoMigrar([]string{"-config", m.rutaConfig, "-a", m.rutaDB}); err != nil {
		t.Fatalf("comandoMigrar() = %v", err)
	}

	destino, err := sqlite.New(m.rutaDB)
	if err != nil {
		t.Fatalf("sqlite.New() = %v", err)
	}

	t.Cleanup(func() { destino.Close() })

	cadena, err := destino.Cadena(ctx, m.tenant)
	if err != nil {
		t.Fatalf("Cadena() = %v", err)
	}

	if len(cadena) != 2 {
		t.Fatalf("Cadena() = %d entradas, want 2", len(cadena))
	}

	if cadena[0].Huella != m.entrada1.Huella || cadena[1].Huella != m.entrada2.Huella {
		t.Errorf("huellas = %q y %q, want %q y %q", cadena[0].Huella, cadena[1].Huella, m.entrada1.Huella, m.entrada2.Huella)
	}

	pendientes, err := destino.Pendientes(ctx, m.tenant, 0)
	if err != nil {
		t.Fatalf("Pendientes() = %v", err)
	}

	if len(pendientes) != 1 {
		t.Fatalf("Pendientes() = %d, want 1", len(pendientes))
	}

	if pendientes[0].Secuencia != m.entrada2.Secuencia {
		t.Errorf("Pendientes()[0] = secuencia %d, want %d", pendientes[0].Secuencia, m.entrada2.Secuencia)
	}

	ultimoEnvio, err := destino.UltimoEnvio(ctx, m.tenant)
	if err != nil {
		t.Fatalf("UltimoEnvio() = %v", err)
	}

	if ultimoEnvio.CSV != m.envio.CSV {
		t.Errorf("UltimoEnvio() CSV = %q, want %q", ultimoEnvio.CSV, m.envio.CSV)
	}

	if !ultimoEnvio.Instante.Equal(m.envio.Instante) {
		t.Errorf("UltimoEnvio() Instante = %v, want %v", ultimoEnvio.Instante, m.envio.Instante)
	}

	if len(ultimoEnvio.Lineas) != 1 {
		t.Errorf("UltimoEnvio() = %d lineas, want 1", len(ultimoEnvio.Lineas))
	}

	vacia, err := destino.Cadena(ctx, m.sinFacturas)
	if err != nil {
		t.Fatalf("Cadena(sin facturas) = %v", err)
	}

	if len(vacia) != 0 {
		t.Errorf("Cadena(sin facturas) = %d entradas, want 0", len(vacia))
	}

	if _, err := destino.UltimoEnvio(ctx, m.sinFacturas); !errors.Is(err, verifactu.ErrNoEncontrado) {
		t.Errorf("UltimoEnvio(sin facturas) = %v, want ErrNoEncontrado", err)
	}

	if err := comandoMigrar([]string{"-config", m.rutaConfig, "-a", m.rutaDB}); err == nil {
		t.Fatal("comandoMigrar() sobre un destino existente = nil, want error")
	}
}

func TestComandoMigrarFalloBorraDestino(t *testing.T) {
	m := prepararMigracion(t)

	rutaCadena := filepath.Join(m.rutaLedger, m.tenant.NIF+"-"+m.tenant.IDSistemaInformatico+".jsonl")

	contenido, err := os.ReadFile(rutaCadena)
	if err != nil {
		t.Fatalf("os.ReadFile() = %v", err)
	}

	if !strings.Contains(string(contenido), m.entrada2.Huella) {
		t.Fatalf("la huella de la segunda entrada no está en %s", rutaCadena)
	}

	manipulado := strings.ReplaceAll(string(contenido), m.entrada2.Huella, strings.Repeat("A", 64))

	if err := os.WriteFile(rutaCadena, []byte(manipulado), 0o600); err != nil {
		t.Fatalf("os.WriteFile() = %v", err)
	}

	if err := comandoMigrar([]string{"-config", m.rutaConfig, "-a", m.rutaDB}); err == nil {
		t.Fatal("comandoMigrar() con la cadena manipulada = nil, want error")
	}

	for _, ruta := range []string{m.rutaDB, m.rutaDB + "-wal", m.rutaDB + "-shm"} {
		if _, err := os.Stat(ruta); !os.IsNotExist(err) {
			t.Errorf("os.Stat(%s) = %v, want que no exista", filepath.Base(ruta), err)
		}
	}
}
