package aeat

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/cristianemek/go-verifactu/record"
)

func TestLeerYDeserializarSobre(t *testing.T) {

	file, err := os.ReadFile("../testdata/xml/oficial/alta-9.1.1.1.xml")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	var s sobre

	err = xml.Unmarshal(file, &s)
	if err != nil {
		t.Fatalf("Error unmarshalling file: %v", err)
	}

	if len(s.Body.Peticion.RegistroFactura) != 1 {
		t.Fatalf("Expected 1 RegistroFactura, but found %d", len(s.Body.Peticion.RegistroFactura))
	}

	if s.Body.Peticion.RegistroFactura[0].RegistroAlta == nil {
		t.Fatalf("RegistroFactura is nil")
	}

	alta := s.Body.Peticion.RegistroFactura[0].RegistroAlta

	if s.Body.Peticion.Cabecera.ObligadoEmision.NombreRazon != "XXXXX" {
		t.Fatalf("Expected NombreRazon to be 'XXXXX', but found '%s'", s.Body.Peticion.Cabecera.ObligadoEmision.NombreRazon)
	}

	if s.Body.Peticion.Cabecera.ObligadoEmision.NIF != "AAAA" {
		t.Fatalf("Expected NIF to be 'AAAA', but found '%s'", s.Body.Peticion.Cabecera.ObligadoEmision.NIF)
	}

	if alta.IDVersion != "1.0" {
		t.Fatalf("Expected IDVersion to be '1.0', but found '%s'", alta.IDVersion)
	}

	if alta.IDFactura.IDEmisorFactura != "AAAA" {
		t.Fatalf("Expected IDEmisorFactura to be 'AAAA', but found '%s'", alta.IDFactura.IDEmisorFactura)
	}

	if alta.IDFactura.NumSerieFactura != "12345" {
		t.Fatalf("Expected NumSerieFactura to be '12345', but found '%s'", alta.IDFactura.NumSerieFactura)
	}

	if alta.IDFactura.FechaExpedicionFactura.Format() != "13-09-2024" {
		t.Fatalf("Expected FechaExpedicionFactura to be '13-09-2024', but found '%s'", alta.IDFactura.FechaExpedicionFactura.Format())
	}

	if alta.TipoFactura != record.TipoFacturaCompleta {
		t.Fatalf("Expected TipoFactura to be 'Completa', but found '%s'", alta.TipoFactura)
	}

	if alta.DescripcionOperacion != "Descripc" {
		t.Fatalf("Expected DescripcionOperacion to be 'Descripc', but found '%s'", alta.DescripcionOperacion)
	}

	if alta.Destinatarios == nil {
		t.Fatalf("Expected Destinatarios to not be nil, but found nil")
	}

	if len(alta.Destinatarios.IDDestinatario) != 1 {
		t.Fatalf("Expected 1 IDDestinatario, but found %d", len(alta.Destinatarios.IDDestinatario))
	}

	if alta.Destinatarios.IDDestinatario[0].NIF == nil {
		t.Fatalf("Expected IDDestinatario NIF to not be nil, but found nil")
	}

	if *alta.Destinatarios.IDDestinatario[0].NIF != "BBBB" {
		t.Fatalf("Expected NIF to be 'BBBB', but found '%s'", *alta.Destinatarios.IDDestinatario[0].NIF)
	}

	if len(alta.Desglose.DetalleDesglose) != 2 {
		t.Fatalf("Expected 2 DetalleDesglose, but found %d", len(alta.Desglose.DetalleDesglose))
	}

	if alta.Desglose.DetalleDesglose[0].ClaveRegimen == nil {
		t.Fatalf("Expected ClaveRegimen to not be nil, but found nil")
	}

	if *alta.Desglose.DetalleDesglose[0].ClaveRegimen != record.ClaveRegimenGeneral {
		t.Fatalf("Expected ClaveRegimen to be '01', but found '%s'", *alta.Desglose.DetalleDesglose[0].ClaveRegimen)
	}

	if alta.Desglose.DetalleDesglose[0].CalificacionOperacion == nil {
		t.Fatalf("Expected CalificacionOperacion to not be nil, but found nil")
	}

	if *alta.Desglose.DetalleDesglose[0].CalificacionOperacion != record.CalificacionOperacionSujetaNoExentaSinISP {
		t.Fatalf("Expected CalificacionOperacion to be 'SujetaNoExentaSinISP', but found '%s'", *alta.Desglose.DetalleDesglose[0].CalificacionOperacion)
	}

	if alta.Desglose.DetalleDesglose[0].TipoImpositivo == nil {
		t.Fatalf("Expected TipoImpositivo to not be nil, but found nil")
	}

	if *alta.Desglose.DetalleDesglose[0].TipoImpositivo != 400 {
		t.Fatalf("Expected TipoImpositivo to be '400', but found '%d'", *alta.Desglose.DetalleDesglose[0].TipoImpositivo)
	}

	if alta.Desglose.DetalleDesglose[0].BaseImponibleOimporteNoSujeto != 1000 {
		t.Fatalf("Expected BaseImponible to be '1000', but found '%d'", alta.Desglose.DetalleDesglose[0].BaseImponibleOimporteNoSujeto)
	}

	if alta.Desglose.DetalleDesglose[0].CuotaRepercutida == nil {
		t.Fatalf("Expected CuotaRepercutida to not be nil, but found nil")
	}

	if *alta.Desglose.DetalleDesglose[0].CuotaRepercutida != 40 {
		t.Fatalf("Expected CuotaRepercutida to be '40', but found '%d'", *alta.Desglose.DetalleDesglose[0].CuotaRepercutida)
	}

	if alta.Desglose.DetalleDesglose[1].TipoImpositivo == nil {
		t.Fatalf("Expected TipoImpositivo to not be nil, but found nil")
	}

	if *alta.Desglose.DetalleDesglose[1].TipoImpositivo != 2100 {
		t.Fatalf("Expected TipoImpositivo to be '2100', but found '%d'", *alta.Desglose.DetalleDesglose[1].TipoImpositivo)
	}

	if alta.Desglose.DetalleDesglose[1].BaseImponibleOimporteNoSujeto != 10000 {
		t.Fatalf("Expected BaseImponibleOimporteNoSujeto to be '10000', but found '%d'", alta.Desglose.DetalleDesglose[1].BaseImponibleOimporteNoSujeto)
	}

	if alta.Desglose.DetalleDesglose[1].CuotaRepercutida == nil {
		t.Fatalf("Expected CuotaRepercutida to not be nil, but found nil")
	}

	if *alta.Desglose.DetalleDesglose[1].CuotaRepercutida != 2100 {
		t.Fatalf("Expected CuotaRepercutida to be '2100', but found '%d'", *alta.Desglose.DetalleDesglose[1].CuotaRepercutida)
	}

	if alta.CuotaTotal != 2140 {
		t.Fatalf("Expected CuotaTotal to be '2140', but found '%d'", alta.CuotaTotal)
	}

	if alta.ImporteTotal != 13140 {
		t.Fatalf("Expected ImporteTotal to be '13140', but found '%d'", alta.ImporteTotal)
	}

	if alta.Encadenamiento.RegistroAnterior == nil {
		t.Fatalf("Expected RegistroAnterior to not be nil, but found nil")
	}

	if alta.Encadenamiento.PrimerRegistro != nil {
		t.Fatalf("Expected PrimerRegistro to be nil, but found not nil")
	}

	if alta.Encadenamiento.RegistroAnterior.IDEmisorFactura != "AAAA" {
		t.Fatalf("Expected RegistroAnterior IDEmisorFactura to be 'AAAA', but found '%s'", alta.Encadenamiento.RegistroAnterior.IDEmisorFactura)
	}

	if alta.Encadenamiento.RegistroAnterior.NumSerieFactura != "44" {
		t.Fatalf("Expected RegistroAnterior NumSerieFactura to be '44', but found '%s'", alta.Encadenamiento.RegistroAnterior.NumSerieFactura)
	}

	if alta.Encadenamiento.RegistroAnterior.FechaExpedicionFactura.Format() != "13-09-2024" {
		t.Fatalf("Expected RegistroAnterior FechaExpedicionFactura to be '13-09-2024', but found '%s'", alta.Encadenamiento.RegistroAnterior.FechaExpedicionFactura.Format())
	}

	if alta.Encadenamiento.RegistroAnterior.Huella != "HuellaRegistroAnterior" {
		t.Fatalf("Expected RegistroAnterior HuellaRegistroAnterior to be equal to Huella, but found '%s' and '%s'", alta.Encadenamiento.RegistroAnterior.Huella, alta.Huella)
	}

	if alta.FechaHoraHusoGenRegistro.Format() != "2024-09-13T19:20:30+01:00" {
		t.Fatalf("Expected FechaHoraHusoGenRegistro to be '2024-09-13T19:20:30+01:00', but found '%s'", alta.FechaHoraHusoGenRegistro.Format())
	}

	if alta.TipoHuella != record.TipoHuellaSHA256 {
		t.Fatalf("Expected TipoHuella to be 'SHA256', but found '%s'", alta.TipoHuella)
	}

	if alta.Huella != "Huella" {
		t.Fatalf("Expected Huella to be '%s', but found '%s'", "Huella", alta.Huella)
	}

	if alta.SistemaInformatico.IdSistemaInformatico != "77" {
		t.Fatalf("Expected IdSistemaInformatico to be '77', but found '%s'", alta.SistemaInformatico.IdSistemaInformatico)
	}

	if alta.SistemaInformatico.Version != "1.0.03" {
		t.Fatalf("Expected Version to be '1.0.03', but found '%s'", alta.SistemaInformatico.Version)
	}

	if alta.SistemaInformatico.NumeroInstalacion != "383" {
		t.Fatalf("Expected NumeroInstalacion to be '383', but found '%s'", alta.SistemaInformatico.NumeroInstalacion)
	}

	if alta.SistemaInformatico.NIF == nil {
		t.Fatalf("Expected NIF to not be nil, but found nil")
	}

	if *alta.SistemaInformatico.NIF != "NNNN" {
		t.Fatalf("Expected NIF to be 'NNNN', but found '%s'", *alta.SistemaInformatico.NIF)
	}

	if alta.SistemaInformatico.TipoUsoPosibleSoloVerifactu != record.SiNoNo {
		t.Fatalf("Expected TipoUsoPosibleSoloVerifactu to be 'No', but found '%s'", alta.SistemaInformatico.TipoUsoPosibleSoloVerifactu)
	}

	if alta.SistemaInformatico.TipoUsoPosibleMultiOT != record.SiNoSi {
		t.Fatalf("Expected TipoUsoPosibleMultiOT to be 'Si', but found '%s'", alta.SistemaInformatico.TipoUsoPosibleMultiOT)
	}

	if alta.SistemaInformatico.IndicadorMultiplesOT != record.SiNoSi {
		t.Fatalf("Expected IndicadorMultiplesOT to be 'Si', but found '%s'", alta.SistemaInformatico.IndicadorMultiplesOT)
	}

}
