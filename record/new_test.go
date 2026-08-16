package record

import (
	"errors"
	"testing"
	"time"
)

func TestNewRegistroAltaRellenaCampos(t *testing.T) {
	rec := validRegistroAlta()
	rec.IDVersion = ""
	rec.TipoHuella = ""
	rec.Huella = ""

	got, err := NewRegistroAlta(rec)
	if err != nil {
		t.Fatalf("NewRegistroAlta returned error: %v", err)
	}

	if got.IDVersion == "" {
		t.Errorf("Expected IDVersion to be filled in, got empty string")
	}

	if got.TipoHuella == "" {
		t.Errorf("Expected TipoHuella to be filled in, got empty string")
	}

	if len(got.Huella) != 64 {
		t.Errorf("Expected Huella to be 64 characters long, got %d", len(got.Huella))
	}
}

func TestNewRegistroAltaHuellaOficial(t *testing.T) {
	fecha := Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	generacion := FechaHora(time.Date(2024, 1, 1, 19, 20, 30, 0, time.FixedZone("CET", 3600)))

	rec := RegistroAlta{
		IDFactura: IDFacturaExpedida{
			IDEmisorFactura:        "89890001K",
			NumSerieFactura:        "12345678/G33",
			FechaExpedicionFactura: fecha,
		},
		NombreRazonEmisor:    "EMPRESA DE PRUEBAS SL",
		TipoFactura:          TipoFacturaCompleta,
		DescripcionOperacion: "Servicios de desarrollo de software",
		Desglose: Desglose{
			DetalleDesglose: []DetalleDesglose{
				{
					Impuesto:                      Ptr(ImpuestoIVA),
					ClaveRegimen:                  Ptr(ClaveRegimenGeneral),
					CalificacionOperacion:         Ptr(CalificacionOperacionSujetaNoExentaSinISP),
					BaseImponibleOimporteNoSujeto: 11100,
					CuotaRepercutida:              Ptr(Amount(1235)),
				},
			},
		},
		CuotaTotal:     1235,
		ImporteTotal:   12345,
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
		FechaHoraHusoGenRegistro: generacion,
	}

	got, err := NewRegistroAlta(rec)
	if err != nil {
		t.Fatalf("NewRegistroAlta returned error: %v", err)
	}

	const want = "3C464DAF61ACB827C65FDA19F352A4E3BDC2C640E9E9FC4CC058073F38F12F60"
	if got.Huella != want {
		t.Errorf("Huella = %q, want %q", got.Huella, want)
	}
}

func TestNewRegistroAltaInvalido(t *testing.T) {
	rec := validRegistroAlta()
	rec.Desglose = Desglose{}

	got, err := NewRegistroAlta(rec)
	if err == nil {
		t.Fatalf("Expected error for invalid RegistroAlta, got nil")
	}

	if !errors.Is(err, ErrValidation) {
		t.Errorf("Expected error to be of type ErrValidation, got %v", err)
	}

	if got.IDVersion != "" {
		t.Errorf("Expected empty RegistroAlta, got %+v", got)
	}
}

func TestNewRegistroAnulacionRellenaCampos(t *testing.T) {
	rec := validRegistroAnulacion()
	rec.IDVersion = ""
	rec.TipoHuella = ""
	rec.Huella = ""

	got, err := NewRegistroAnulacion(rec)
	if err != nil {
		t.Fatalf("NewRegistroAnulacion returned error: %v", err)
	}

	if got.IDVersion == "" {
		t.Errorf("Expected IDVersion to be filled in, got empty string")
	}

	if got.TipoHuella == "" {
		t.Errorf("Expected TipoHuella to be filled in, got empty string")
	}

	if len(got.Huella) != 64 {
		t.Errorf("Expected Huella to be 64 characters long, got %d", len(got.Huella))
	}
}

func TestNewRegistroAnulacionHuellaOficial(t *testing.T) {
	fecha := Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	generacion := FechaHora(time.Date(2024, 1, 1, 19, 20, 40, 0, time.FixedZone("CET", 3600)))

	enc := NewEncadenamientoRegistroAnterior("89890001K", "12345679/G34", fecha, "F7B94CFD8924EDFF273501B01EE5153E4CE8F259766F88CF6ACB8935802A2B97")

	rec := RegistroAnulacion{
		IDFactura: IDFacturaExpedidaBaja{
			IDEmisorFacturaAnulada:        "89890001K",
			NumSerieFacturaAnulada:        "12345679/G34",
			FechaExpedicionFacturaAnulada: fecha,
		},
		Encadenamiento: enc,
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
		FechaHoraHusoGenRegistro: generacion,
	}

	got, err := NewRegistroAnulacion(rec)
	if err != nil {
		t.Fatalf("NewRegistroAnulacion returned error: %v", err)
	}

	const want = "177547C0D57AC74748561D054A9CEC14B4C4EA23D1BEFD6F2E69E3A388F90C68"
	if got.Huella != want {
		t.Errorf("Huella = %q, want %q", got.Huella, want)
	}
}

func TestNewRegistroAnulacionInvalido(t *testing.T) {
	rec := validRegistroAnulacion()
	rec.Encadenamiento = Encadenamiento{}

	got, err := NewRegistroAnulacion(rec)
	if err == nil {
		t.Fatalf("Expected error for invalid RegistroAnulacion, got nil")
	}

	if !errors.Is(err, ErrValidation) {
		t.Errorf("Expected error to be of type ErrValidation, got %v", err)
	}

	if got.IDVersion != "" {
		t.Errorf("Expected empty RegistroAnulacion, got %+v", got)
	}
}
