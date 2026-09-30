package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu/record"
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

func TestCargarConfigInvalida(t *testing.T) {
	tenant := `"tenants": {"89890001K": {"token": "t", "certificado": "c", "tipo_certificado": "representante"}}`

	testCases := []struct {
		name   string
		config string
	}{
		{name: "json mal formado", config: `{`},
		{name: "sin listen", config: `{"data": "d", "entorno": "pruebas", ` + tenant + `}`},
		{name: "sin data", config: `{"listen": ":1", "entorno": "pruebas", ` + tenant + `}`},
		{name: "sin entorno", config: `{"listen": ":1", "data": "d", ` + tenant + `}`},
		{name: "sin tenants", config: `{"listen": ":1", "data": "d", "entorno": "pruebas"}`},
		{name: "sin tipo_certificado", config: `{"listen": ":1", "data": "d", "entorno": "pruebas", "tenants": {"89890001K": {"token": "t", "certificado": "c"}}}`},
		{name: "NIF repetido en minusculas", config: `{"listen": ":1", "data": "d", "entorno": "pruebas", "tenants": {
			"89890001K": {"token": "t", "certificado": "c", "tipo_certificado": "representante"},
			"89890001k": {"token": "u", "certificado": "c", "tipo_certificado": "representante"}}}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tc.config), 0o600); err != nil {
				t.Fatalf("WriteFile() = %v", err)
			}

			if _, err := cargarConfig(path); err == nil {
				t.Error("cargarConfig() = nil, want error")
			}
		})
	}

	if _, err := cargarConfig(filepath.Join(t.TempDir(), "no-existe.json")); err == nil {
		t.Error("cargarConfig() de un fichero que no existe = nil, want error")
	}
}

func certificadoDePrueba(t *testing.T) string {
	t.Helper()

	clave, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() = %v", err)
	}

	plantilla := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "prueba"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}

	der, err := x509.CreateCertificate(rand.Reader, plantilla, plantilla, &clave.PublicKey, clave)
	if err != nil {
		t.Fatalf("CreateCertificate() = %v", err)
	}

	derClave, err := x509.MarshalPKCS8PrivateKey(clave)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey() = %v", err)
	}

	pemCompleto := append(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: derClave})...,
	)

	ruta := filepath.Join(t.TempDir(), "certificado.pem")
	if err := os.WriteFile(ruta, pemCompleto, 0o600); err != nil {
		t.Fatalf("WriteFile() = %v", err)
	}

	return ruta
}

func TestConstruirServidor(t *testing.T) {
	certificado := certificadoDePrueba(t)

	configDe := func(store TipoStore, data, cert, tipo string) *Config {
		return &Config{
			Entorno: "pruebas",
			Store:   store,
			Data:    data,
			Sistema: record.SistemaInformatico{IdSistemaInformatico: "01"},
			Tenants: map[string]TenantConfig{
				"89890001K": {Nombre: "EMPRESA DE PRUEBAS SL", Token: "t", Certificado: cert, TipoCertificado: tipo},
			},
		}
	}

	testCases := []struct {
		name    string
		cfg     func(dir string) *Config
		wantErr bool
	}{
		{
			name: "ledger",
			cfg: func(dir string) *Config {
				return configDe(StoreLedger, filepath.Join(dir, "datos"), certificado, "representante")
			},
		},
		{
			name: "sqlite",
			cfg: func(dir string) *Config {
				return configDe(StoreSQLite, filepath.Join(dir, "verifactu.db"), certificado, "representante")
			},
		},
		{
			name: "ledger sobre un fichero",
			cfg: func(dir string) *Config {
				ruta := filepath.Join(dir, "fichero")
				os.WriteFile(ruta, nil, 0o600)
				return configDe(StoreLedger, ruta, certificado, "representante")
			},
			wantErr: true,
		},
		{
			name: "sqlite sobre un directorio",
			cfg: func(dir string) *Config {
				return configDe(StoreSQLite, dir, certificado, "representante")
			},
			wantErr: true,
		},
		{
			name: "sin certificado",
			cfg: func(dir string) *Config {
				return configDe(StoreLedger, filepath.Join(dir, "datos"), filepath.Join(dir, "no-existe.pem"), "representante")
			},
			wantErr: true,
		},
		{
			name: "tipo de certificado desconocido",
			cfg: func(dir string) *Config {
				return configDe(StoreLedger, filepath.Join(dir, "datos"), certificado, "otro")
			},
			wantErr: true,
		},
		{
			name: "cadena rota",
			cfg: func(dir string) *Config {
				datos := filepath.Join(dir, "datos")
				os.MkdirAll(datos, 0o755)
				os.WriteFile(filepath.Join(datos, "89890001K-01.jsonl"), []byte(`{"Secuencia":2,"Operacion":"alta"}`+"\n"), 0o600)
				return configDe(StoreLedger, datos, certificado, "representante")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := construirServidor(tc.cfg(t.TempDir()))

			if tc.wantErr {
				if err == nil {
					srv.cerrar()
					t.Fatal("construirServidor() = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("construirServidor() = %v", err)
			}

			if err := srv.cerrar(); err != nil {
				t.Errorf("cerrar() = %v", err)
			}
		})
	}
}

func TestEjecutar(t *testing.T) {
	certificado := certificadoDePrueba(t)

	anterior := slog.Default()
	t.Cleanup(func() { slog.SetDefault(anterior) })

	escribir := func(t *testing.T, listen, remision, log, cert string) string {
		t.Helper()

		dir := t.TempDir()
		config := fmt.Sprintf(`{
			"listen": %q,
			"data": %q,
			"entorno": "pruebas",
			"remision_cada": %q,
			"log": %q,
			"sistema": {"IdSistemaInformatico": "01"},
			"tenants": {"89890001K": {"nombre": "EMPRESA DE PRUEBAS SL", "token": "t", "certificado": %q, "tipo_certificado": "representante"}}
		}`, listen, filepath.Join(dir, "datos"), remision, log, cert)

		ruta := filepath.Join(dir, "config.json")
		if err := os.WriteFile(ruta, []byte(config), 0o600); err != nil {
			t.Fatalf("WriteFile() = %v", err)
		}
		return ruta
	}

	testCases := []struct {
		name    string
		args    func(t *testing.T) []string
		wantErr bool
	}{
		{
			name: "arranca y para",
			args: func(t *testing.T) []string {
				return []string{"-config", escribir(t, "127.0.0.1:0", "60s", filepath.Join(t.TempDir(), "verifactud.log"), certificado)}
			},
		},
		{
			name:    "flag desconocido",
			args:    func(t *testing.T) []string { return []string{"-otro"} },
			wantErr: true,
		},
		{
			name: "config que no existe",
			args: func(t *testing.T) []string {
				return []string{"-config", filepath.Join(t.TempDir(), "no-existe.json")}
			},
			wantErr: true,
		},
		{
			name: "remision_cada invalido",
			args: func(t *testing.T) []string {
				return []string{"-config", escribir(t, "127.0.0.1:0", "cada rato", "", certificado)}
			},
			wantErr: true,
		},
		{
			name: "log en un directorio",
			args: func(t *testing.T) []string {
				return []string{"-config", escribir(t, "127.0.0.1:0", "60s", t.TempDir(), certificado)}
			},
			wantErr: true,
		},
		{
			name: "sin certificado",
			args: func(t *testing.T) []string {
				return []string{"-config", escribir(t, "127.0.0.1:0", "60s", "", filepath.Join(t.TempDir(), "no-existe.pem"))}
			},
			wantErr: true,
		},
		{
			name: "no puede escuchar",
			args: func(t *testing.T) []string {
				return []string{"-config", escribir(t, "sin-puerto", "60s", "", certificado)}
			},
			wantErr: true,
		},
		{
			name:    "migrar sin destino",
			args:    func(t *testing.T) []string { return []string{"migrar"} },
			wantErr: true,
		},
		{
			name: "migrar",
			args: func(t *testing.T) []string {
				m := prepararMigracion(t)
				return []string{"migrar", "-config", m.rutaConfig, "-a", m.rutaDB}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			err := ejecutar(ctx, tc.args(t))

			if (err != nil) != tc.wantErr {
				t.Errorf("ejecutar() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestComandoMigrarErrores(t *testing.T) {
	testCases := []struct {
		name string
		args func(t *testing.T) []string
	}{
		{
			name: "flag desconocido",
			args: func(t *testing.T) []string { return []string{"-otro"} },
		},
		{
			name: "config que no existe",
			args: func(t *testing.T) []string {
				return []string{"-config", filepath.Join(t.TempDir(), "no-existe.json"), "-a", filepath.Join(t.TempDir(), "x.db")}
			},
		},
		{
			name: "el origen no es ledger",
			args: func(t *testing.T) []string {
				m := prepararMigracion(t)
				cambiarConfig(t, m.rutaConfig, `"store": "ledger"`, `"store": "sqlite"`)
				return []string{"-config", m.rutaConfig, "-a", m.rutaDB}
			},
		},
		{
			name: "el ledger no se puede abrir",
			args: func(t *testing.T) []string {
				m := prepararMigracion(t)
				fichero := filepath.Join(t.TempDir(), "fichero")
				os.WriteFile(fichero, nil, 0o600)
				cambiarConfig(t, m.rutaConfig, fmt.Sprintf("%q", m.rutaLedger), fmt.Sprintf("%q", fichero))
				return []string{"-config", m.rutaConfig, "-a", m.rutaDB}
			},
		},
		{
			name: "el destino no se puede crear",
			args: func(t *testing.T) []string {
				m := prepararMigracion(t)
				return []string{"-config", m.rutaConfig, "-a", filepath.Join(t.TempDir(), "no", "existe", "x.db")}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := comandoMigrar(tc.args(t)); err == nil {
				t.Error("comandoMigrar() = nil, want error")
			}
		})
	}
}

func cambiarConfig(t *testing.T, ruta, viejo, nuevo string) {
	t.Helper()

	contenido, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatalf("ReadFile() = %v", err)
	}

	if !strings.Contains(string(contenido), viejo) {
		t.Fatalf("%s no contiene %s", ruta, viejo)
	}

	if err := os.WriteFile(ruta, []byte(strings.Replace(string(contenido), viejo, nuevo, 1)), 0o600); err != nil {
		t.Fatalf("WriteFile() = %v", err)
	}
}
