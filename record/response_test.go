package record

import (
	"encoding/xml"
	"os"
	"testing"
	"time"
)

func TestParseo(t *testing.T) {
	data, err := os.ReadFile("../testdata/xml/respuesta/respuesta-correcto-no-oficial.xml")
	if err != nil {
		t.Fatalf("Error reading XML file: %v", err)
	}

	var res RespuestaRegFactuSistemaFacturacion

	err = xml.Unmarshal(data, &res)
	if err != nil {
		t.Fatalf("Error unmarshaling XML: %v", err)
	}

	if res.CSV != "A1B2C3D4E5F6G7H8" {
		t.Errorf("Expected CSV 'A1B2C3D4E5F6G7H8', got '%s'", res.CSV)
	}

	if res.EstadoEnvio != EstadoEnvioCorrecto {
		t.Errorf("Expected EstadoEnvio 'Correcto', got '%s'", res.EstadoEnvio)
	}

	if res.TiempoEsperaEnvio != "60" {
		t.Errorf("Expected TiempoEsperaEnvio '60', got '%s'", res.TiempoEsperaEnvio)
	}

	if res.Cabecera.ObligadoEmision.NIF != "89890001K" {
		t.Errorf("Expected NIF '89890001K', got '%s'", res.Cabecera.ObligadoEmision.NIF)
	}

	if res.DatosPresentacion == nil {
		t.Fatalf("Expected DatosPresentacion to be non-nil")
	}

	if res.DatosPresentacion.NIFPresentador != "89890001K" {
		t.Errorf("Expected NIFPresentador '89890001K', got '%s'", res.DatosPresentacion.NIFPresentador)
	}

	if res.DatosPresentacion.TimestampPresentacion.Format(time.RFC3339) != "2024-01-01T19:20:30+01:00" {
		t.Errorf("Expected TimestampPresentacion '2024-01-01T19:20:30+01:00', got '%s'", res.DatosPresentacion.TimestampPresentacion.Format(time.RFC3339))
	}

	if len(res.RespuestaLinea) != 1 {
		t.Fatalf("Expected 1 RespuestaLinea, got %d", len(res.RespuestaLinea))
	}

	if res.RespuestaLinea[0].IDFactura.NumSerieFactura != "001" {
		t.Errorf("Expected NumSerieFactura '001', got '%s'", res.RespuestaLinea[0].IDFactura.NumSerieFactura)
	}

	if res.RespuestaLinea[0].IDFactura.FechaExpedicionFactura.Format() != "01-01-2024" {
		t.Errorf("Expected FechaExpedicionFactura '01-01-2024', got '%s'", res.RespuestaLinea[0].IDFactura.FechaExpedicionFactura.Format())
	}

	if res.RespuestaLinea[0].EstadoRegistro != EstadoRegistroCorrecto {
		t.Errorf("Expected EstadoRegistro 'Correcto', got '%s'", res.RespuestaLinea[0].EstadoRegistro)
	}

	if res.RespuestaLinea[0].RegistroDuplicado != nil || res.RespuestaLinea[0].Operacion.Subsanacion != nil {
		t.Errorf("Expected RegistroDuplicado and Subsanacion to be nil, got RegistroDuplicado: %+v, Subsanacion: %+v", res.RespuestaLinea[0].RegistroDuplicado, res.RespuestaLinea[0].Operacion.Subsanacion)
	}

	if res.RespuestaLinea[0].Operacion.TipoOperacion != TipoOperacionAlta {
		t.Errorf("Expected TipoOperacion 'Alta', got '%s'", res.RespuestaLinea[0].Operacion.TipoOperacion)
	}

}

func TestRegistroParcialmenteCorrecto(t *testing.T) {
	data, err := os.ReadFile("../testdata/xml/respuesta/respuesta-parcialmente-correcto-no-oficial.xml")
	if err != nil {
		t.Fatalf("Error reading XML file: %v", err)
	}

	var res RespuestaRegFactuSistemaFacturacion

	err = xml.Unmarshal(data, &res)
	if err != nil {
		t.Fatalf("Error unmarshaling XML: %v", err)
	}

	if res.EstadoEnvio != EstadoEnvioParcialmenteCorrecto {
		t.Errorf("Expected EstadoEnvio 'ParcialmenteCorrecto', got '%s'", res.EstadoEnvio)
	}

	if len(res.RespuestaLinea) != 2 {
		t.Fatalf("Expected 2 RespuestaLinea, got %d", len(res.RespuestaLinea))
	}

	respuestaLinea0 := res.RespuestaLinea[0]

	if respuestaLinea0.EstadoRegistro != EstadoRegistroCorrecto {
		t.Errorf("Expected EstadoRegistro 'Correcto' for first RespuestaLinea, got '%s'", respuestaLinea0.EstadoRegistro)
	}

	if respuestaLinea0.CodigoErrorRegistro != "" {
		t.Errorf("Expected CodigoErrorRegistro to be empty for first RespuestaLinea, got '%s'", respuestaLinea0.CodigoErrorRegistro)
	}

	if respuestaLinea0.RefExterna != "" {
		t.Errorf("Expected RefExterna to be empty for first RespuestaLinea, got '%s'", respuestaLinea0.RefExterna)
	}

	respuestaLinea1 := res.RespuestaLinea[1]

	if respuestaLinea1.EstadoRegistro != EstadoRegistroAceptadoConErrores {
		t.Errorf("Expected EstadoRegistro 'AceptadoConErrores' for second RespuestaLinea, got '%s'", respuestaLinea1.EstadoRegistro)
	}

	if respuestaLinea1.IDFactura.NumSerieFactura != "002" {
		t.Errorf("Expected NumSerieFactura '002' for second RespuestaLinea, got '%s'", respuestaLinea1.IDFactura.NumSerieFactura)
	}

	if respuestaLinea1.CodigoErrorRegistro != "2000" {
		t.Errorf("Expected CodigoErrorRegistro '2000' for second RespuestaLinea, got '%s'", respuestaLinea1.CodigoErrorRegistro)
	}

	if respuestaLinea1.DescripcionErrorRegistro == "" {
		t.Errorf("Expected DescripcionErrorRegistro to be non-empty for second RespuestaLinea")
	}

	if respuestaLinea1.RefExterna != "PEDIDO-4711" {
		t.Errorf("Expected RefExterna 'PEDIDO-4711' for second RespuestaLinea, got '%s'", respuestaLinea1.RefExterna)
	}

}

func TestRegistroIncorrecto(t *testing.T) {
	data, err := os.ReadFile("../testdata/xml/respuesta/respuesta-incorrecto-no-oficial.xml")
	if err != nil {
		t.Fatalf("Error reading XML file: %v", err)
	}

	var res RespuestaRegFactuSistemaFacturacion

	err = xml.Unmarshal(data, &res)
	if err != nil {
		t.Fatalf("Error unmarshaling XML: %v", err)
	}

	if res.EstadoEnvio != EstadoEnvioIncorrecto {
		t.Errorf("Expected EstadoEnvio 'Incorrecto', got '%s'", res.EstadoEnvio)
	}

	if res.CSV != "" {
		t.Errorf("Expected CSV to be empty, got '%s'", res.CSV)
	}

	if res.DatosPresentacion != nil {
		t.Errorf("Expected DatosPresentacion to be nil, got %+v", res.DatosPresentacion)
	}

	if len(res.RespuestaLinea) != 1 {
		t.Fatalf("Expected 1 RespuestaLinea, got %d", len(res.RespuestaLinea))
	}

	if res.RespuestaLinea[0].EstadoRegistro != EstadoRegistroIncorrecto {
		t.Errorf("Expected EstadoRegistro 'Incorrecto' for RespuestaLinea, got '%s'", res.RespuestaLinea[0].EstadoRegistro)
	}

	if res.RespuestaLinea[0].CodigoErrorRegistro != "1106" {
		t.Errorf("Expected CodigoErrorRegistro '1106' for RespuestaLinea, got '%s'", res.RespuestaLinea[0].CodigoErrorRegistro)
	}

	if res.RespuestaLinea[0].RegistroDuplicado != nil {
		t.Errorf("Expected RegistroDuplicado to be nil, got %+v", res.RespuestaLinea[0].RegistroDuplicado)
	}

}

func TestRegistroDuplicado(t *testing.T) {
	data, err := os.ReadFile("../testdata/xml/respuesta/respuesta-duplicado-no-oficial.xml")
	if err != nil {
		t.Fatalf("Error reading XML file: %v", err)
	}

	var res RespuestaRegFactuSistemaFacturacion

	err = xml.Unmarshal(data, &res)
	if err != nil {
		t.Fatalf("Error unmarshaling XML: %v", err)
	}

	if len(res.RespuestaLinea) != 1 {
		t.Fatalf("Expected 1 RespuestaLinea, got %d", len(res.RespuestaLinea))
	}

	respuestaLinea := res.RespuestaLinea[0]

	if respuestaLinea.RegistroDuplicado == nil {
		t.Fatalf("Expected RegistroDuplicado to not be nil, got %+v", respuestaLinea.RegistroDuplicado)
	}

	if respuestaLinea.CodigoErrorRegistro != "3000" {
		t.Errorf("Expected CodigoErrorRegistro '3000', got '%s'", respuestaLinea.CodigoErrorRegistro)
	}

	if respuestaLinea.RegistroDuplicado.EstadoRegistroDuplicado != EstadoRegistroAlmacenadoCorrecta {
		t.Errorf("Expected EstadoRegistroDuplicado 'Correcta', got '%s'", respuestaLinea.RegistroDuplicado.EstadoRegistroDuplicado)
	}

	if respuestaLinea.RegistroDuplicado.IdPeticionRegistroDuplicado != "20240101192030ABCD" {
		t.Errorf("Expected IdPeticionRegistroDuplicado '20240101192030ABCD', got '%s'", respuestaLinea.RegistroDuplicado.IdPeticionRegistroDuplicado)
	}

}
