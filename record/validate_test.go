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

func validRegistroAnulacion() RegistroAnulacion {
	return RegistroAnulacion{
		IDVersion: "1.0",
		IDFactura: IDFacturaExpedidaBaja{
			IDEmisorFacturaAnulada:        "89890001K",
			NumSerieFacturaAnulada:        "12345678/G33",
			FechaExpedicionFacturaAnulada: Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
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
		FechaHoraHusoGenRegistro: FechaHora(time.Date(2024, 1, 1, 19, 20, 40, 0, time.FixedZone("CET", 3600))),
		TipoHuella:               TipoHuellaSHA256,
		Huella:                   "177547C0D57AC74748561D054A9CEC14B4C4EA23D1BEFD6F2E69E3A388F90C68",
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

func TestValidateRequiredFieldsRegistroAlta(t *testing.T) {
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
		{
			name: "NombreRazonEmisor is longer than 120 characters",
			mod: func(r *RegistroAlta) {
				r.NombreRazonEmisor = strings.Repeat("a", 121)
			},
		},
		{
			name: "Huella is too long",
			mod: func(r *RegistroAlta) {
				r.Huella = strings.Repeat("a", 65)
			},
		},
		{
			name: "NIF is too long",
			mod: func(r *RegistroAlta) {
				r.SistemaInformatico.NIF = Ptr("1234567890")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAlta()
			tc.mod(&rec)
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for missing required field, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		})
	}
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAlta()
			rec.SistemaInformatico = tc.det
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for invalid SistemaInformatico, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		})
	}
}

func TestValidateDestinatarios(t *testing.T) {
	testCases := []struct {
		name string
		mod  func(*RegistroAlta)
	}{
		{
			name: "Destinatarios.IDDestinatario has both NIF and IDOtro set",
			mod: func(r *RegistroAlta) {
				r.Destinatarios = &Destinatarios{
					IDDestinatario: []PersonaFisicaJuridica{
						{
							NIF:    Ptr("12345678A"),
							IDOtro: &IDOtro{},
						},
					},
				}
			},
		},
		{
			name: "Destinatarios.IDDestinatario has neither NIF nor IDOtro set",
			mod: func(r *RegistroAlta) {
				r.Destinatarios = &Destinatarios{
					IDDestinatario: []PersonaFisicaJuridica{{}},
				}
			},
		},
		{
			name: "Destinatarios.IDDestinatario has invalid NIF",
			mod: func(r *RegistroAlta) {
				r.Destinatarios = &Destinatarios{
					IDDestinatario: []PersonaFisicaJuridica{
						{
							NIF: Ptr(strings.Repeat("a", 8)),
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAlta()
			tc.mod(&rec)
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for invalid Destinatarios, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		})
	}
}

func TestValidRegistroAnulacion(t *testing.T) {
	rec := validRegistroAnulacion()
	if err := rec.Validate(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidateRequiredFieldsRegistroAnulacion(t *testing.T) {
	testCases := []struct {
		name string
		mod  func(*RegistroAnulacion)
	}{
		{
			name: "Encadenamiento is empty",
			mod: func(r *RegistroAnulacion) {
				r.Encadenamiento = Encadenamiento{}
			},
		},
		{
			name: "Encadenamiento has both PrimerRegistro and RegistroAnterior set",
			mod: func(r *RegistroAnulacion) {
				r.Encadenamiento = Encadenamiento{
					PrimerRegistro:   Ptr(PrimerRegistroCadenaSi),
					RegistroAnterior: Ptr(EncadenamientoFacturaAnterior{}),
				}
			},
		},
		{
			name: "IDVersion is empty",
			mod: func(r *RegistroAnulacion) {
				r.IDVersion = ""
			},
		},
		{
			name: "TipoHuella is empty",
			mod: func(r *RegistroAnulacion) {
				r.TipoHuella = ""
			},
		},
		{
			name: "Huella is not 64 characters long",
			mod: func(r *RegistroAnulacion) {
				r.Huella = "short"
			},
		},
		{
			name: "Huella is longer than 64 characters",
			mod: func(r *RegistroAnulacion) {
				r.Huella = strings.Repeat("a", 65)
			},
		},
		{
			name: "Huella is shorter than 64 characters",
			mod: func(r *RegistroAnulacion) {
				r.Huella = strings.Repeat("a", 63)
			},
		},
		{
			name: "NIF is too short",
			mod: func(r *RegistroAnulacion) {
				r.SistemaInformatico.NIF = Ptr("12345678")
			},
		},
		{
			name: "NIF is too long",
			mod: func(r *RegistroAnulacion) {
				r.SistemaInformatico.NIF = Ptr("1234567890")
			},
		},
		{
			name: "NumSerieFacturaAnulada is empty",
			mod: func(r *RegistroAnulacion) {
				r.IDFactura.NumSerieFacturaAnulada = ""
			},
		},
		{
			name: "NumSerieFacturaAnulada is longer than 60 characters",
			mod: func(r *RegistroAnulacion) {
				r.IDFactura.NumSerieFacturaAnulada = strings.Repeat("a", 61)
			},
		},
		{
			name: "SistemaInformatico has both NIF and IDOtro set",
			mod: func(r *RegistroAnulacion) {
				r.SistemaInformatico.NIF = Ptr("12345678A")
				r.SistemaInformatico.IDOtro = &IDOtro{}
			},
		},
		{
			name: "IdSistemaInformatico is longer than 2 characters",
			mod: func(r *RegistroAnulacion) {
				r.SistemaInformatico.IdSistemaInformatico = "123"
			},
		},
		{
			name: "SistemaInformatico has neither NIF nor IDOtro set",
			mod: func(r *RegistroAnulacion) {
				r.SistemaInformatico.NIF = nil
				r.SistemaInformatico.IDOtro = nil
			},
		},
		{
			name: "IDEmisorFacturaAnulada is invalid",
			mod: func(r *RegistroAnulacion) {
				r.IDFactura.IDEmisorFacturaAnulada = "12345678"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAnulacion()
			tc.mod(&rec)
			err := rec.Validate()
			if err == nil {
				t.Fatalf("Expected error for missing required field, got nil")
			}

			if !errors.Is(err, ErrValidation) {
				t.Errorf("Expected error to be of type ErrValidation, got %v", err)
			}
		})
	}
}

func TestValidateVariantesValidas(t *testing.T) {
	testCases := []struct {
		name string
		mod  func(*RegistroAlta)
	}{
		{
			name: "SistemaInformatico identified with IDOtro instead of NIF",
			mod: func(r *RegistroAlta) {
				r.SistemaInformatico.NIF = nil
				r.SistemaInformatico.IDOtro = &IDOtro{
					CodigoPais: Ptr("FR"),
					IDType:     PersonaFisicaJuridicaIDTypePasaporte,
					ID:         "X1234567",
				}
			},
		},
		{
			name: "Destinatarios is nil",
			mod: func(r *RegistroAlta) {
				r.Destinatarios = nil
			},
		},
		{
			name: "DetalleDesglose with OperacionExenta instead of CalificacionOperacion",
			mod: func(r *RegistroAlta) {
				r.Desglose.DetalleDesglose = []DetalleDesglose{
					{
						Impuesto:                      Ptr(ImpuestoIVA),
						ClaveRegimen:                  Ptr(ClaveRegimenGeneral),
						OperacionExenta:               Ptr(OperacionExentaE1),
						BaseImponibleOimporteNoSujeto: 10000,
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAlta()
			tc.mod(&rec)
			err := rec.Validate()
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
