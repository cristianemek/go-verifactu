package record

import (
	"encoding/xml"
	"testing"
	"time"
)

const (
	xmlPrimerRegistro   = "<Encadenamiento><PrimerRegistro>S</PrimerRegistro></Encadenamiento>"
	xmlRegistroAnterior = "<Encadenamiento><RegistroAnterior><IDEmisorFactura>89890001K</IDEmisorFactura><NumSerieFactura>12345678/G33</NumSerieFactura><FechaExpedicionFactura>01-01-2024</FechaExpedicionFactura><Huella>3C464DAF61ACB827C65FDA19F352A4E3BDC2C640E9E9FC4CC058073F38F12F60</Huella></RegistroAnterior></Encadenamiento>"
)

func TestEncadenamientoPrimerRegistroMarshalXML(t *testing.T) {

	enc := NewEncadenamientoPrimerRegistro()

	xmlEnc, err := xml.Marshal(enc)

	if err != nil {
		t.Fatalf("Encadenamiento.MarshalXML returned error: %v", err)
	}

	if string(xmlEnc) != xmlPrimerRegistro {
		t.Errorf("Encadenamiento.MarshalXML() = %q, want %q", string(xmlEnc), xmlPrimerRegistro)
	}
}

func TestEncadenamientoRegistroAnteriorMarshalXML(t *testing.T) {

	enc := NewEncadenamientoRegistroAnterior("89890001K", "12345678/G33", Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)), "3C464DAF61ACB827C65FDA19F352A4E3BDC2C640E9E9FC4CC058073F38F12F60")
	xmlEnc, err := xml.Marshal(enc)

	if err != nil {
		t.Fatalf("Encadenamiento.MarshalXML returned error: %v", err)
	}

	if string(xmlEnc) != xmlRegistroAnterior {
		t.Errorf("Encadenamiento.MarshalXML() = %q, want %q", string(xmlEnc), xmlRegistroAnterior)
	}
}
