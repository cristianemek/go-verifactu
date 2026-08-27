package aeat

import (
	"encoding/xml"
	"strings"
	"testing"

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
