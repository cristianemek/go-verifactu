package record

import (
	"encoding/xml"
	"os"
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

func TestRegistroAltaMarshalXMLShape(t *testing.T) {
	expedicion, err := time.Parse(fechaFormat, "01-01-2024")
	if err != nil {
		t.Fatalf("Error parsing expedition date: %v", err)
	}

	generacion, err := time.Parse(time.RFC3339, "2024-01-01T19:20:30+01:00")
	if err != nil {
		t.Fatalf("Error parsing generation date: %v", err)
	}

	rec := RegistroAlta{
		IDVersion: "1.0",
		IDFactura: IDFacturaExpedida{
			IDEmisorFactura:        "89890001K",
			NumSerieFactura:        "12345678/G33",
			FechaExpedicionFactura: Fecha(expedicion),
		},
		NombreRazonEmisor:    "EMPRESA DE PRUEBAS SL",
		TipoFactura:          TipoFacturaCompleta,
		DescripcionOperacion: "Servicios de desarrollo de software",
		Destinatarios: &Destinatarios{
			IDDestinatario: []PersonaFisicaJuridica{
				{NombreRazon: "CLIENTE SL", NIF: Ptr("B12345674")},
			},
		},
		Desglose: Desglose{
			DetalleDesglose: []DetalleDesglose{
				{
					Impuesto:                      Ptr(ImpuestoIVA),
					ClaveRegimen:                  Ptr(ClaveRegimenGeneral),
					CalificacionOperacion:         Ptr(CalificacionOperacionSujetaNoExentaSinISP),
					TipoImpositivo:                Ptr(Porcentaje(2100)),
					BaseImponibleOimporteNoSujeto: 10000,
					CuotaRepercutida:              Ptr(Amount(2100)),
				},
			},
		},
		CuotaTotal:     2100,
		ImporteTotal:   12100,
		Encadenamiento: NewEncadenamientoPrimerRegistro(),
		SistemaInformatico: SistemaInformatico{
			NombreRazon:                 "MI EMPRESA SL",
			NIF:                         Ptr("B87654321"),
			NombreSistemaInformatico:    "go-verifactu",
			IdSistemaInformatico:        "01",
			Version:                     "0.1.0",
			NumeroInstalacion:           "0001",
			TipoUsoPosibleSoloVerifactu: SiNoSi,
			TipoUsoPosibleMultiOT:       SiNoNo,
			IndicadorMultiplesOT:        SiNoNo,
		},
		FechaHoraHusoGenRegistro: FechaHora(generacion),
		TipoHuella:               TipoHuellaSHA256,
		Huella:                   "3C464DAF61ACB827C65FDA19F352A4E3BDC2C640E9E9FC4CC058073F38F12F60",
	}

	envelope := RegFactuSistemaFacturacion{
		Cabecera: Cabecera{
			ObligadoEmision: PersonaFisicaJuridicaES{
				NombreRazon: "EMPRESA DE PRUEBAS SL",
				NIF:         "89890001K",
			},
		},
		RegistroFactura: []RegistroFactura{
			{RegistroAlta: &rec},
		},
	}

	out, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent returned error: %v", err)
	}

	t.Logf("\n%s", out)

	err = os.MkdirAll("../testdata/xml", 0755)
	if err != nil {
		t.Fatalf("Error creating test data directory: %v", err)
	}

	err = os.WriteFile("../testdata/xml/registro_alta.xml", out, 0644)
	if err != nil {
		t.Fatalf("Error writing test data file: %v", err)
	}
}
