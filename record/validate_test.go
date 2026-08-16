package record

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validRegistroAlta() RegistroAlta {
	return RegistroAlta{
		IDVersion: "1.0",
		IDFactura: IDFacturaExpedida{
			IDEmisorFactura:        "89890001K",
			NumSerieFactura:        "12345678/G33",
			FechaExpedicionFactura: Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
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
		FechaHoraHusoGenRegistro: FechaHora(time.Date(2024, 1, 1, 19, 20, 30, 0, time.FixedZone("CET", 3600))),
		TipoHuella:               TipoHuellaSHA256,
		Huella:                   "3C464DAF61ACB827C65FDA19F352A4E3BDC2C640E9E9FC4CC058073F38F12F60",
	}
}

func TestValidateValido(t *testing.T) {
	rec := validRegistroAlta()
	if err := rec.Validate(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidateEncadenamientoInvalido(t *testing.T) {
	testCases := []struct {
		name string
		enc  Encadenamiento
	}{
		{
			name: "Both PrimerRegistro and RegistroAnterior are nil",
			enc:  Encadenamiento{},
		},
		{
			name: "Both PrimerRegistro and RegistroAnterior are set",
			enc: Encadenamiento{
				PrimerRegistro:   Ptr(PrimerRegistroCadenaSi),
				RegistroAnterior: Ptr(EncadenamientoFacturaAnterior{}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAlta()
			rec.Encadenamiento = tc.enc
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for invalid Encadenamiento, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		})
	}
}

func TestValidateDesgloseInvalido(t *testing.T) {
	var tooLongDetalleDesglose []DetalleDesglose
	for i := 0; i < 13; i++ {
		tooLongDetalleDesglose = append(tooLongDetalleDesglose, DetalleDesglose{
			CalificacionOperacion: Ptr(CalificacionOperacionSujetaNoExentaSinISP),
		})
	}

	testCases := []struct {
		name string
		desg Desglose
	}{
		{
			name: "Empty Desglose",
			desg: Desglose{},
		},
		{
			name: "Too many DetalleDesglose entries",
			desg: Desglose{
				DetalleDesglose: tooLongDetalleDesglose,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAlta()
			rec.Desglose = tc.desg
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for invalid Desglose, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		})
	}
}

func TestValidateDetalleDesgloseChoice(t *testing.T) {
	testCases := []struct {
		name string
		det  DetalleDesglose
	}{
		{
			name: "Both CalificacionOperacion and OperacionExenta are nil",
			det:  DetalleDesglose{},
		},
		{
			name: "Both CalificacionOperacion and OperacionExenta are set",
			det: DetalleDesglose{
				CalificacionOperacion: Ptr(CalificacionOperacionSujetaNoExentaSinISP),
				OperacionExenta:       Ptr(OperacionExentaE1),
			},
		},
	}

	t.Run("DetalleDesglose choice validation", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				rec := validRegistroAlta()
				rec.Desglose.DetalleDesglose = []DetalleDesglose{tc.det}
				err := rec.Validate()
				if err == nil {
					t.Fatalf("Expected error for invalid DetalleDesglose choice, got nil")
				}

				if !errors.Is(err, ErrValidation) {
					t.Errorf("Expected error to be of type ErrValidation, got %v", err)
				}
			})
		}

	})

}

func TestValidateRequiredFields(t *testing.T) {
	testCases := []struct {
		name string
		mod  func(*RegistroAlta)
	}{
		{
			name: "NombreRazonEmisor is empty",
			mod: func(r *RegistroAlta) {
				r.NombreRazonEmisor = ""
			},
		},
		{
			name: "IDVersion is empty",
			mod: func(r *RegistroAlta) {
				r.IDVersion = ""
			},
		},
		{
			name: "DescripcionOperacion is empty",
			mod: func(r *RegistroAlta) {
				r.DescripcionOperacion = ""
			},
		},
		{
			name: "TipoHuella is empty",
			mod: func(r *RegistroAlta) {
				r.TipoHuella = ""
			},
		},
		{
			name: "Huella is not 64 characters long",
			mod: func(r *RegistroAlta) {
				r.Huella = "short"
			},
		},
		{
			name: "SistemaInformatico.IdSistemaInformatico is longer than 2 characters",
			mod: func(r *RegistroAlta) {
				r.SistemaInformatico.IdSistemaInformatico = "123"
			},
		},
		{
			name: "DescripcionOperacion is longer than 500 characters",
			mod: func(r *RegistroAlta) {
				r.DescripcionOperacion = strings.Repeat("a", 501)
			},
		},
		{
			name: "IDEmisorFactura must contain 9 characters",
			mod: func(r *RegistroAlta) {
				r.IDFactura.IDEmisorFactura = "12345678"
			},
		},
		{
			name: "NumSerieFactura is empty",
			mod: func(r *RegistroAlta) {
				r.IDFactura.NumSerieFactura = ""
			},
		},
		{
			name: "NumSerieFactura is longer than 60 characters",
			mod: func(r *RegistroAlta) {
				r.IDFactura.NumSerieFactura = strings.Repeat("a", 61)
			},
		},
	}

	t.Run("Required fields validation", func(t *testing.T) {
		for _, tc := range testCases {
			rec := validRegistroAlta()
			tc.mod(&rec)
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for missing required field, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		}
	})
}

func TestValidateSistemaInformatico(t *testing.T) {
	testCases := []struct {
		name string
		det  SistemaInformatico
	}{
		{
			name: "Both NIF and IDOtro are nil",
			det:  SistemaInformatico{},
		},
		{
			name: "Both NIF and IDOtro are set",
			det: SistemaInformatico{
				NIF:    Ptr(""),
				IDOtro: Ptr(IDOtro{}),
			},
		},
	}

	t.Run("SistemaInformatico validation", func(t *testing.T) {
		for _, tc := range testCases {
			rec := validRegistroAlta()
			rec.SistemaInformatico = tc.det
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for invalid SistemaInformatico, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		}
	})
}
