package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/memory"
)

const (
	facturaJSON = `{
  "IDFactura": {
    "IDEmisorFactura": "89890001K",
    "NumSerieFactura": "F-2026-001",
    "FechaExpedicionFactura": "10-09-2026"
  },
  "NombreRazonEmisor": "EMPRESA DE PRUEBAS SL",
  "TipoFactura": "F2",
  "DescripcionOperacion": "Servicios de desarrollo",
  "Desglose": {
    "DetalleDesglose": [
      {
        "Impuesto": "01",
        "ClaveRegimen": "01",
        "CalificacionOperacion": "S1",
        "TipoImpositivo": 2100,
        "BaseImponibleOimporteNoSujeto": 10000,
        "CuotaRepercutida": 2100
      }
    ]
  },
  "CuotaTotal": 2100,
  "ImporteTotal": 12100
}
`
)

func decodificar(t *testing.T, rec *httptest.ResponseRecorder) respuestaRegistro {
	t.Helper()

	var resp respuestaRegistro
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("Error decoding response: %v, body: %s", err, rec.Body.String())
	}

	return resp
}

func peticionGET(s *servidor, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/89890001K/estado?"+query, nil)

	req.SetPathValue("nif", "89890001K")
	req.Header.Set("Authorization", "Bearer secreto")

	rec := httptest.NewRecorder()

	s.auth(s.estado).ServeHTTP(rec, req)

	return rec
}

func servidorConStore(t *testing.T) (*servidor, *memory.Store) {
	t.Helper()

	store := memory.New()

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

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: func() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) }, SistemaInformatico: &sistema})

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	return &servidor{
		engine:  engine,
		tenants: map[string]TenantConfig{"89890001K": {Nombre: "EMPRESA DE PRUEBAS SL", Token: "secreto"}},
		sistema: "01",
		entorno: record.EntornoPruebas,
	}, store
}

func peticion(s *servidor, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/89890001K/alta", strings.NewReader(cuerpo))

	req.SetPathValue("nif", "89890001K")
	req.Header.Set("Authorization", "Bearer secreto")
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	s.auth(s.alta).ServeHTTP(rec, req)

	return rec

}

func TestServidor(t *testing.T) {

	testCases := []struct {
		name           string
		nif            string
		header         string
		expectedStatus int
	}{
		{
			name:           "Valid tenant and token",
			nif:            "89890001K",
			header:         "Bearer valid_token",
			expectedStatus: 204,
		},
		{
			name:           "NIF desconocido",
			nif:            "00000000X",
			header:         "Bearer valid_token",
			expectedStatus: 401,
		},
		{
			name:           "Invalid token",
			nif:            "89890001K",
			header:         "Bearer invalid_token",
			expectedStatus: 401,
		},
		{
			name:           "Missing token",
			nif:            "89890001K",
			header:         "",
			expectedStatus: 401,
		},
		{
			name:           "NIF en minúscula",
			nif:            "89890001k",
			header:         "Bearer valid_token",
			expectedStatus: 204,
		},
	}
	s := &servidor{
		tenants: map[string]TenantConfig{
			"89890001K": {Token: "valid_token"},
		},
	}

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/"+tc.nif+"/alta", nil)

			req.SetPathValue("nif", tc.nif)

			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			rec := httptest.NewRecorder()

			s.auth(next).ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}

}

func TestAlta(t *testing.T) {
	s, _ := servidorConStore(t)

	testCases := []struct {
		name           string
		cuerpo         string
		expectedStatus int
	}{
		{
			name:           "Valid invoice",
			cuerpo:         facturaJSON,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Json mal formado",
			cuerpo:         `{"x":`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Factura sin desglose",
			cuerpo:         `{"IDFactura":{"IDEmisorFactura":"89890001K","NumSerieFactura":"F-2026-009","FechaExpedicionFactura":"10-09-2026"}}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := peticion(s, tc.cuerpo)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d, body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}

			resp := decodificar(t, rec)

			if !strings.HasPrefix(resp.QR, "https://prewww2.aeat.es") && tc.expectedStatus == http.StatusCreated {
				t.Errorf("Expected QR URL to start with AEAT test URL, got %s", resp.QR)
			}

		})

	}
}

func TestAltaIdempotente(t *testing.T) {
	s, _ := servidorConStore(t)

	rec := peticion(s, facturaJSON)

	rec2 := peticion(s, facturaJSON)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	if rec2.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d, body: %s", http.StatusCreated, rec2.Code, rec2.Body.String())
	}

	resp := decodificar(t, rec)

	if resp.Entry == nil {
		t.Fatalf("Expected entry, got nil")
	}

	if resp.Entry.Secuencia != 1 {
		t.Fatalf("Expected sequence 1, got %d", resp.Entry.Secuencia)
	}

	resp2 := decodificar(t, rec2)

	if resp2.Entry == nil {
		t.Fatalf("Expected entry, got nil")
	}

	if resp2.Entry.Secuencia != 1 {
		t.Fatalf("Expected sequence 1, got %d", resp2.Entry.Secuencia)
	}

	if resp.Entry.Huella != resp2.Entry.Huella {
		t.Fatalf("Expected same fingerprint, got %s and %s", resp.Entry.Huella, resp2.Entry.Huella)
	}

}

func TestAltaConAvisos(t *testing.T) {
	s, _ := servidorConStore(t)

	factura := strings.Replace(facturaJSON, `"BaseImponibleOimporteNoSujeto": 10000`, `"BaseImponibleOimporteNoSujeto": 0`, 1)

	rec := peticion(s, factura)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	resp := decodificar(t, rec)

	if resp.Entry == nil {
		t.Fatalf("Expected entry, got nil")
	}

	if len(resp.Avisos) != 1 {
		t.Fatalf("Expected 1 aviso, got %d", len(resp.Avisos))
	}

	if !strings.Contains(resp.Avisos[0], "2005") {
		t.Fatalf("Expected aviso about base imponible, got %s", resp.Avisos[0])
	}
}

func TestEstado(t *testing.T) {
	s, _ := servidorConStore(t)

	rec := peticion(s, facturaJSON)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	testCases := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "existente",
			query:          "serie=F-2026-001&fecha=10-09-2026",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "sin serie",
			query:          "fecha=10-09-2026",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "fecha en ISO",
			query:          "serie=F-2026-001&fecha=2026-09-10",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "op invalida",
			query:          "serie=F-2026-001&fecha=10-09-2026&op=foo",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no existe",
			query:          "serie=F-2026-999&fecha=10-09-2026",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "anulacion inexistente",
			query:          "serie=F-2026-001&fecha=10-09-2026&op=anulacion",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := peticionGET(s, tc.query)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d, body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}

			if tc.expectedStatus == http.StatusOK {
				resp := decodificar(t, rec)
				if resp.Entry == nil {
					t.Fatalf("Expected entry, got nil")
				}

				if resp.Entry.Secuencia != 1 {
					t.Fatalf("Expected sequence 1, got %d", resp.Entry.Secuencia)
				}
			}

		})
	}
}

func TestEstadoAEAT(t *testing.T) {

	testCases := []struct {
		name   string
		envio  *verifactu.Envio
		estado string
		codigo string
		csv    string
	}{
		{
			name:   "pendiente",
			envio:  nil,
			estado: "Pendiente",
		},
		{
			name: "correcta",
			envio: &verifactu.Envio{
				Lineas: []verifactu.LineaEnvio{
					{
						Estado:    record.EstadoRegistroCorrecto,
						Secuencia: 1,
					},
				},
				CSV: "A-1",
			},
			estado: "Correcto",
			csv:    "A-1",
		},
		{
			name: "rechazada",
			envio: &verifactu.Envio{
				Lineas: []verifactu.LineaEnvio{
					{Secuencia: 1, Estado: record.EstadoRegistroIncorrecto, CodigoError: "1189", Descripcion: "Faltan destinatarios"},
				},
				CSV: "A-1",
			},
			estado: "Incorrecto",
			codigo: "1189",
			csv:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, store := servidorConStore(t)

			rec := peticion(s, facturaJSON)

			if rec.Code != http.StatusCreated {
				t.Fatalf("Expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
			}

			if tc.envio != nil {
				tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}
				if err := store.AnexarEnvio(context.Background(), tenant, tc.envio); err != nil {
					t.Fatalf("Error anexando envio: %v", err)
				}

			}

			rec = peticionGET(s, "serie=F-2026-001&fecha=10-09-2026")
			resp := decodificar(t, rec)

			if resp.AEAT == nil {
				t.Fatalf("Expected aeat block, got nil")
			}

			if resp.AEAT.Estado != tc.estado {
				t.Errorf("Expected estado %s, got %s", tc.estado, resp.AEAT.Estado)
			}
			if resp.AEAT.Codigo != tc.codigo {
				t.Errorf("Expected codigo %s, got %s", tc.codigo, resp.AEAT.Codigo)
			}
			if resp.AEAT.CSV != tc.csv {
				t.Errorf("Expected csv %s, got %s", tc.csv, resp.AEAT.CSV)
			}

		})
	}

}
