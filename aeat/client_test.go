package aeat

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

type redirigirA struct {
	destino *url.URL
}

func (r *redirigirA) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	req.URL.Scheme = r.destino.Scheme
	req.URL.Host = r.destino.Host

	return http.DefaultTransport.RoundTrip(req)
}

func clienteContra(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(h)

	t.Cleanup(srv.Close)

	destino, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("Error al parsear la URL del servidor de prueba: %v", err)
	}

	c, err := NewClient(Config{
		Entorno:         EntornoPruebas,
		TipoCertificado: CertificadoRepresentante,
		HTTPClient: &http.Client{
			Transport: &redirigirA{destino: destino},
		},
	})

	if err != nil {
		t.Fatalf("Error al crear el cliente: %v", err)
	}

	return c, srv
}

func TestClientRemitirEnviaLaPeticion(t *testing.T) {

	respuesta, err := os.ReadFile("../testdata/xml/sobre/sobre-respuesta-correcta-no-oficial.xml")
	if err != nil {
		t.Fatalf("Error al leer el archivo: %v", err)
	}

	var capturado struct {
		metodo      string
		ruta        string
		contentType string
		userAgent   string
		body        []byte
		soapAction  string
	}

	c, _ := clienteContra(t, func(w http.ResponseWriter, r *http.Request) {
		capturado.metodo = r.Method
		capturado.ruta = r.URL.Path
		capturado.contentType = r.Header.Get("Content-Type")
		capturado.userAgent = r.Header.Get("User-Agent")
		capturado.soapAction = r.Header.Get("SOAPAction")
		capturado.body, _ = io.ReadAll(r.Body)

		w.Write(respuesta)
	})

	resp, err := c.Remitir(context.Background(), verifactu.Tenant{}, record.RegFactuSistemaFacturacion{})
	if err != nil {
		t.Fatalf("Error al remitir: %v", err)
	}

	if capturado.metodo != http.MethodPost {
		t.Errorf("Método HTTP esperado %s, obtenido %s", http.MethodPost, capturado.metodo)
	}

	if capturado.ruta != "/wlpl/TIKE-CONT/ws/SistemaFacturacion/VerifactuSOAP" {
		t.Errorf("Ruta esperada %s, obtenida %s", "/wlpl/TIKE-CONT/ws/SistemaFacturacion/VerifactuSOAP", capturado.ruta)

	}

	if capturado.contentType != "text/xml; charset=utf-8" {
		t.Errorf("Content-Type esperado %s, obtenido %s", "text/xml; charset=utf-8", capturado.contentType)
	}

	if capturado.userAgent != ("go-verifactu/" + verifactu.Version) {
		t.Errorf("User-Agent esperado %s, obtenido %s", "go-verifactu/"+verifactu.Version, capturado.userAgent)
	}

	if !bytes.Contains(capturado.body, []byte("RegFactuSistemaFacturacion")) {
		t.Errorf("El cuerpo de la petición no contiene un sobre SOAP válido")
	}

	if resp.CSV != "A1B2C3D4E5F6G7H8" {
		t.Errorf("CSV esperado %s, obtenido %s", "A1B2C3D4E5F6G7H8", resp.CSV)
	}

}

// A 500 status with an internal fault is a retry indicator, not an HTTP error.
func TestClientRemitirElEstadoHTTPNoDecide(t *testing.T) {
	testCases := []struct {
		name          string
		fixture       string
		status        int
		expectedError error
	}{
		{
			name:          "500 con fallo interno",
			fixture:       "../testdata/xml/sobre/sobre-fault-servidor-no-oficial.xml",
			status:        http.StatusInternalServerError,
			expectedError: verifactu.ErrFaultServidor,
		},
		{
			name:          "200 con fallo interno",
			fixture:       "../testdata/xml/sobre/sobre-fault-servidor-no-oficial.xml",
			status:        http.StatusOK,
			expectedError: verifactu.ErrFaultServidor,
		},
		{
			name:          "sobre respuesta correcta",
			fixture:       "../testdata/xml/sobre/sobre-respuesta-correcta-no-oficial.xml",
			status:        http.StatusInternalServerError,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			respuesta, err := os.ReadFile(tc.fixture)
			if err != nil {
				t.Fatalf("Error al leer el archivo: %v", err)
			}

			c, _ := clienteContra(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				w.Write(respuesta)
			})

			resp, err := c.Remitir(context.Background(), verifactu.Tenant{}, record.RegFactuSistemaFacturacion{})

			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Fatalf("Error inesperado: %v", err)
				}

				var sf *SoapFault

				if !errors.As(err, &sf) {
					t.Fatalf("Se esperaba un error de tipo SoapFault, pero se obtuvo: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Error inesperado: %v", err)
				}

				if resp.CSV != "A1B2C3D4E5F6G7H8" {
					t.Errorf("CSV esperado %s, obtenido %s", "A1B2C3D4E5F6G7H8", resp.CSV)
				}
			}

		})

	}
}

func TestClientRemitirErrorDeTransporte(t *testing.T) {
	c, srv := clienteContra(t, func(w http.ResponseWriter, r *http.Request) {})

	srv.Close()

	_, err := c.Remitir(context.Background(), verifactu.Tenant{}, record.RegFactuSistemaFacturacion{})
	if err == nil {
		t.Fatalf("Se esperaba un error de transporte, pero no se obtuvo ninguno")
	}

}

func TestNewClientSinCertificado(t *testing.T) {
	client, err := NewClient(Config{
		Entorno:         EntornoPruebas,
		TipoCertificado: CertificadoRepresentante,
	})
	if !errors.Is(err, ErrCertificadoRequerido) {
		t.Fatalf("Se esperaba un error de certificado, pero no se obtuvo ninguno")
	}

	if client != nil {
		t.Fatalf("Se esperaba un cliente nulo, pero se obtuvo uno no nulo")
	}

}

func TestNewClientConHTTPClientNoExigeCertificado(t *testing.T) {
	_, err := NewClient(Config{
		Entorno:         EntornoPruebas,
		TipoCertificado: CertificadoRepresentante,
		HTTPClient:      &http.Client{},
	})

	if err != nil {
		t.Fatalf("Error al crear el cliente: %v", err)
	}

}
