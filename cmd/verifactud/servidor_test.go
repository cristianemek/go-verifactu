package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/aeat"
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

	return servidorConTransporte(t, nil)
}

func servidorConTransporte(t *testing.T, tp verifactu.Transport) (*servidor, *memory.Store) {
	t.Helper()

	store := memory.New()

	return servidorSobre(t, store, tp), store
}

func servidorSobre(t *testing.T, store verifactu.Store, tp verifactu.Transport) *servidor {
	t.Helper()

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

	engine, err := verifactu.New(verifactu.Config{Store: store, Transport: tp, Now: func() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) }, SistemaInformatico: &sistema})

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	return &servidor{
		engine:  engine,
		tenants: map[string]TenantConfig{"89890001K": {Nombre: "EMPRESA DE PRUEBAS SL", Token: "secreto"}},
		sistema: "01",
		entorno: record.EntornoPruebas,
		avisar:  make(chan struct{}, 1),
		almacen: store,
	}
}

func peticion(s *servidor, cuerpo string) *httptest.ResponseRecorder {
	return peticionAlta(s, "", cuerpo)
}

func peticionAlta(s *servidor, query, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/89890001K/alta?"+query, strings.NewReader(cuerpo))

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
		{
			name:           "Emisor distinto del NIF de la ruta",
			cuerpo:         strings.Replace(facturaJSON, `"IDEmisorFactura": "89890001K"`, `"IDEmisorFactura": "89890002L"`, 1),
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

const anulacionJSON = `{
  "IDFactura": {
    "IDEmisorFacturaAnulada": "89890001K",
    "NumSerieFacturaAnulada": "F-2026-001",
    "FechaExpedicionFacturaAnulada": "10-09-2026"
  }
}`

func peticionAnular(s *servidor, query, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/89890001K/anular?"+query, strings.NewReader(cuerpo))

	req.SetPathValue("nif", "89890001K")
	req.Header.Set("Authorization", "Bearer secreto")

	rec := httptest.NewRecorder()

	s.auth(s.anulacion).ServeHTTP(rec, req)

	return rec
}

func TestAnulacion(t *testing.T) {
	s, _ := servidorConStore(t)

	if rec := peticion(s, facturaJSON); rec.Code != http.StatusCreated {
		t.Fatalf("alta = %d, body: %s", rec.Code, rec.Body.String())
	}

	testCases := []struct {
		name   string
		query  string
		cuerpo string
		want   int
	}{
		{name: "valida", cuerpo: anulacionJSON, want: http.StatusCreated},
		{name: "json mal formado", cuerpo: `{"x":`, want: http.StatusBadRequest},
		{name: "subsanacion no aplica", query: "subsanacion", cuerpo: anulacionJSON, want: http.StatusBadRequest},
		{name: "tras rechazo", query: "tras_rechazo", cuerpo: anulacionJSON, want: http.StatusCreated},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if rec := peticionAnular(s, tc.query, tc.cuerpo); rec.Code != tc.want {
				t.Errorf("anular = %d, want %d, body: %s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}

	rec := peticionGET(s, "serie=F-2026-001&fecha=10-09-2026&op=anulacion")
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d, body: %s", rec.Code, rec.Body.String())
	}

	if resp := decodificar(t, rec); resp.QR != "" {
		t.Errorf("QR de una anulacion = %q, want vacio", resp.QR)
	}
}

func TestQRNoASCII(t *testing.T) {
	s, _ := servidorConStore(t)

	entry := &verifactu.Entry{Alta: &record.RegistroAlta{IDFactura: record.IDFacturaExpedida{IDEmisorFactura: "89890001K", NumSerieFactura: "F-Ñ"}}}

	if qr := s.qrDe(entry); qr != "" {
		t.Errorf("qrDe() = %q, want vacio", qr)
	}
}

func TestEntornoQR(t *testing.T) {
	if got := entornoQR("produccion"); got != record.EntornoProduccion {
		t.Errorf("entornoQR(produccion) = %q", got)
	}

	if got := entornoQR("pruebas"); got != record.EntornoPruebas {
		t.Errorf("entornoQR(pruebas) = %q", got)
	}
}

func TestHealthzYCerrar(t *testing.T) {
	s, _ := servidorConStore(t)

	rec := httptest.NewRecorder()
	s.healthz(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("healthz = %d, want 200", rec.Code)
	}

	if err := s.cerrar(); err != nil {
		t.Errorf("cerrar() con un store sin Close = %v, want nil", err)
	}
}

func TestMapearError(t *testing.T) {
	testCases := []struct {
		err  error
		want int
	}{
		{&verifactu.ErrorEspera{}, http.StatusTooManyRequests},
		{verifactu.ErrSinPendientes, http.StatusOK},
		{record.ErrValidation, http.StatusBadRequest},
		{verifactu.ErrOpcionNoAplicable, http.StatusBadRequest},
		{verifactu.ErrEmisorDistinto, http.StatusBadRequest},
		{verifactu.ErrNoEncontrado, http.StatusNotFound},
		{verifactu.ErrFaultCliente, http.StatusUnprocessableEntity},
		{verifactu.ErrFaultServidor, http.StatusBadGateway},
		{aeat.ErrAccesoDenegado, http.StatusServiceUnavailable},
		{verifactu.ErrConflictoDeSecuencia, http.StatusConflict},
		{context.DeadlineExceeded, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			if got, _ := mapearError(tc.err); got != tc.want {
				t.Errorf("mapearError() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRemitir(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name         string
		tf           *transporteFalso
		altas        []string
		nif          string
		wantErr      bool
		wantLlamadas int32
	}{
		{name: "sin pendientes", tf: &transporteFalso{}, nif: "89890001K"},
		{name: "remite", tf: &transporteFalso{}, altas: []string{"F-2026-001"}, nif: "89890001K", wantLlamadas: 1},
		{name: "rechazado por la cadena", tf: &transporteFalso{estado: record.EstadoRegistroAceptadoConErrores, codigo: "2007"}, altas: []string{"F-2026-001"}, nif: "89890001K", wantLlamadas: 1},
		{name: "rechazado por los datos", tf: &transporteFalso{estado: record.EstadoRegistroIncorrecto, codigo: "1189"}, altas: []string{"F-2026-001"}, nif: "89890001K", wantLlamadas: 1},
		{name: "en espera", tf: &transporteFalso{}, altas: []string{"F-2026-001", "F-2026-002"}, nif: "89890001K", wantLlamadas: 1},
		{name: "fault del cliente", tf: &transporteFalso{err: verifactu.ErrFaultCliente}, altas: []string{"F-2026-001"}, nif: "89890001K", wantErr: true, wantLlamadas: 1},
		{name: "tenant desconocido", tf: &transporteFalso{}, nif: "89890002L", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := servidorConTransporte(t, tc.tf)

			var err error

			for _, serie := range tc.altas {
				if rec := peticion(s, strings.Replace(facturaJSON, "F-2026-001", serie, 1)); rec.Code != http.StatusCreated {
					t.Fatalf("alta = %d, body: %s", rec.Code, rec.Body.String())
				}

				err = s.remitir(ctx, tc.nif)
			}

			if len(tc.altas) == 0 {
				err = s.remitir(ctx, tc.nif)
			}

			if (err != nil) != tc.wantErr {
				t.Fatalf("remitir() = %v, wantErr %v", err, tc.wantErr)
			}

			if got := tc.tf.llamadas.Load(); got != tc.wantLlamadas {
				t.Errorf("llamadas al transporte = %d, want %d", got, tc.wantLlamadas)
			}
		})
	}
}

func TestVueltaBloqueaTrasFaultDelCliente(t *testing.T) {
	testCases := []struct {
		name          string
		err           error
		wantBloqueado bool
		wantLlamadas  int32
	}{
		{name: "fault del cliente", err: verifactu.ErrFaultCliente, wantBloqueado: true, wantLlamadas: 1},
		{name: "otro error", err: context.DeadlineExceeded, wantLlamadas: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tf := &transporteFalso{err: tc.err}
			s, _ := servidorConTransporte(t, tf)

			if rec := peticion(s, facturaJSON); rec.Code != http.StatusCreated {
				t.Fatalf("alta = %d", rec.Code)
			}

			bloqueados := map[string]bool{}

			s.vuelta(bloqueados)
			s.vuelta(bloqueados)

			if bloqueados["89890001K"] != tc.wantBloqueado {
				t.Errorf("bloqueado = %v, want %v", bloqueados["89890001K"], tc.wantBloqueado)
			}

			if got := tf.llamadas.Load(); got != tc.wantLlamadas {
				t.Errorf("llamadas = %d, want %d", got, tc.wantLlamadas)
			}
		})
	}
}

func TestBucleRemision(t *testing.T) {
	tf := &transporteFalso{}
	s, _ := servidorConTransporte(t, tf)

	if rec := peticion(s, facturaJSON); rec.Code != http.StatusCreated {
		t.Fatalf("alta = %d", rec.Code)
	}

	ctx, cancel := context.WithCancel(context.Background())
	hecho := make(chan struct{})

	go func() {
		s.bucleRemision(ctx, time.Millisecond)
		close(hecho)
	}()

	limite := time.Now().Add(2 * time.Second)
	for tf.llamadas.Load() == 0 && time.Now().Before(limite) {
		time.Sleep(time.Millisecond)
	}

	s.tocarTimbre()
	time.Sleep(20 * time.Millisecond)

	cancel()
	<-hecho

	if got := tf.llamadas.Load(); got != 1 {
		t.Errorf("llamadas = %d, want 1", got)
	}
}

type haciaServidor struct {
	destino *url.URL
}

func (h haciaServidor) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = h.destino.Scheme
	req.URL.Host = h.destino.Host
	return http.DefaultTransport.RoundTrip(req)
}

func TestConexion(t *testing.T) {
	faultCliente, err := os.ReadFile("../../testdata/xml/sobre/sobre-fault-cliente-no-oficial.xml")
	if err != nil {
		t.Fatalf("ReadFile() = %v", err)
	}

	testCases := []struct {
		name    string
		nif     string
		handler http.HandlerFunc
		want    int
	}{
		{
			name: "certificado aceptado",
			nif:  "89890001K",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(faultCliente)
			},
			want: http.StatusNoContent,
		},
		{
			name: "certificado rechazado",
			nif:  "89890001K",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "https://sede.agenciatributaria.gob.es/Sede/errores/erro4033.html")
				w.WriteHeader(http.StatusFound)
			},
			want: http.StatusServiceUnavailable,
		},
		{
			name:    "respuesta ilegible",
			nif:     "89890001K",
			handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("no es xml")) },
			want:    http.StatusInternalServerError,
		},
		{
			name:    "sin cliente para ese NIF",
			nif:     "89890002L",
			handler: func(w http.ResponseWriter, r *http.Request) {},
			want:    http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			t.Cleanup(srv.Close)

			destino, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatalf("url.Parse() = %v", err)
			}

			cliente, err := aeat.NewClient(aeat.Config{
				Entorno:         aeat.EntornoPruebas,
				TipoCertificado: aeat.CertificadoRepresentante,
				HTTPClient:      &http.Client{Transport: haciaServidor{destino: destino}},
			})
			if err != nil {
				t.Fatalf("NewClient() = %v", err)
			}

			s, _ := servidorConStore(t)
			s.clientes = map[string]*aeat.Client{"89890001K": cliente}

			req := httptest.NewRequest(http.MethodGet, "/v1/"+tc.nif+"/conexion", nil)
			req.SetPathValue("nif", tc.nif)

			rec := httptest.NewRecorder()
			s.conexion(rec, req)

			if rec.Code != tc.want {
				t.Errorf("conexion = %d, want %d, body: %s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}
