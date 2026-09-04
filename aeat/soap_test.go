package aeat

import (
	"encoding/xml"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

func Test_serializarSobre(t *testing.T) {
	r := record.RegFactuSistemaFacturacion{
		Cabecera: record.Cabecera{
			ObligadoEmision: record.PersonaFisicaJuridicaES{NombreRazon: "Cristian", NIF: "89890001K"},
		},
		RegistroFactura: []record.RegistroFactura{{RegistroAlta: &record.RegistroAlta{}}},
	}

	serialized, err := serializarSobre(r)
	if err != nil {
		t.Fatalf("serializarSobre() error = %v", err)
	}

	if !strings.HasPrefix(string(serialized), xml.Header) {
		t.Errorf("serializarSobre() = %v, want prefix %v", string(serialized), xml.Header)
	}

	if !strings.Contains(string(serialized), nsSOAP) {
		t.Errorf("serializarSobre() = %v, want contain %v", string(serialized), nsSOAP)
	}

	var recuperado sobre

	err = xml.Unmarshal(serialized, &recuperado)
	if err != nil {
		t.Fatalf("xml.Unmarshal() error = %v", err)
	}

	if recuperado.Body.Peticion.Cabecera.ObligadoEmision.NIF != "89890001K" {
		t.Errorf("recuperado.Body.Peticion.Cabecera.ObligadoEmision.NIF = %v, want %v", recuperado.Body.Peticion.Cabecera.ObligadoEmision.NIF, "89890001K")
	}

	if len(recuperado.Body.Peticion.RegistroFactura) != 1 {
		t.Errorf("len(recuperado.Body.Peticion.RegistroFactura) = %v, want %v", len(recuperado.Body.Peticion.RegistroFactura), 1)
	}
}

func TestParsearRespuesta(t *testing.T) {
	testCases := []struct {
		name      string
		file      string
		body      []byte
		wantErr   error
		wantFault bool
		wantCSV   string
	}{
		{
			name:    "respuesta correcta",
			file:    "../testdata/xml/sobre/sobre-respuesta-correcta-no-oficial.xml",
			wantErr: nil,
			wantCSV: "A1B2C3D4E5F6G7H8",
		},
		{
			name:      "fault servidor",
			file:      "../testdata/xml/sobre/sobre-fault-servidor-no-oficial.xml",
			wantErr:   verifactu.ErrFaultServidor,
			wantFault: true,
		},
		{
			name:      "fault cliente",
			file:      "../testdata/xml/sobre/sobre-fault-cliente-no-oficial.xml",
			wantErr:   verifactu.ErrFaultCliente,
			wantFault: true,
		},
		{
			name:    "sobre vacío",
			file:    "",
			body:    []byte(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"> <soapenv:Body> </soapenv:Body> </soapenv:Envelope> `),
			wantErr: ErrRespuestaInesperada,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var body []byte

			if tc.file != "" {
				fileBody, err := os.ReadFile(tc.file)
				if err != nil {
					t.Fatalf("os.ReadFile() error = %v", err)
				}

				body = fileBody
			} else {
				body = tc.body
			}

			respuesta, err := parsearRespuesta(body)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("parsearRespuesta() error = %v, wantErr %v", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("parsearRespuesta() unexpected error = %v", err)
				}
			}

			if tc.wantFault {
				var sf *SoapFault
				if !errors.As(err, &sf) {
					t.Fatalf("parsearRespuesta() error = %v, wantErr %v", err, tc.wantErr)
				}

				if sf.Message == "" {
					t.Errorf("parsearRespuesta() fault message is empty, want non-empty")
				}
			}

			if tc.wantCSV != "" {
				if respuesta.CSV != tc.wantCSV {
					t.Errorf("parsearRespuesta() CSV = %v, want %v", respuesta.CSV, tc.wantCSV)
				}

				if len(respuesta.RespuestaLinea) != 1 {
					t.Errorf("parsearRespuesta() len(respuesta.RespuestaLinea) = %v, want %v", len(respuesta.RespuestaLinea), 1)
				}
			}

		})
	}
}
func TestParsearRespuestaCuerpoInvalido(t *testing.T) {
	_, err := parsearRespuesta([]byte("<html>error 502</html>"))

	if err == nil {
		t.Fatalf("parsearRespuesta() error = %v, wantErr %v", err, ErrRespuestaInesperada)
	}
}
