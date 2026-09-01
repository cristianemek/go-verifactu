package verifactu_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/memory"
)

func respuestaLineaPara(e *verifactu.Entry, tipo record.TipoOperacion, estado record.EstadoRegistro) record.RespuestaLinea {
	return record.RespuestaLinea{
		IDFactura: record.IDFacturaExpedida{
			IDEmisorFactura:        e.IDFactura.NIF,
			NumSerieFactura:        e.IDFactura.NumSerie,
			FechaExpedicionFactura: e.Alta.IDFactura.FechaExpedicionFactura,
		},
		Operacion:      record.OperacionRespuesta{TipoOperacion: tipo},
		EstadoRegistro: estado,
	}
}

func TestRemitirSinTransport(t *testing.T) {
	store := memory.New()

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	_, err = engine.Remitir(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrTransportRequerido) {
		t.Fatalf("Expected error %v, got %v", verifactu.ErrTransportRequerido, err)
	}

}

func TestRemitirSinPendientes(t *testing.T) {
	store := memory.New()

	tf := &transporteFalso{}

	engine, err := verifactu.New(verifactu.Config{Store: store, Transport: tf, Now: fixedTime})

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	_, err = engine.Remitir(context.Background(), tenant)
	if !errors.Is(err, verifactu.ErrSinPendientes) {
		t.Fatalf("Expected error %v, got %v", verifactu.ErrSinPendientes, err)
	}

	if tf.llamadas != 0 {
		t.Errorf("Expected 0 calls to transporteFalso, got %d", tf.llamadas)
	}

}

func TestRemitirCaminoFeliz(t *testing.T) {
	store := memory.New()
	tf := &transporteFalso{}

	engine, err := verifactu.New(verifactu.Config{Store: store, Transport: tf, Now: fixedTime})

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	entry1, err := engine.Alta(context.Background(), tenant, validRegistroAlta("001"))

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	entry2, err := engine.Alta(context.Background(), tenant, validRegistroAlta("002"))

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	entry3, err := engine.Alta(context.Background(), tenant, validRegistroAlta("003"))

	if err != nil {
		t.Fatalf("Error creating engine: %v", err)
	}

	tf.respuesta = record.RespuestaRegFactuSistemaFacturacion{
		CSV:               "CSV-DE-PRUEBA",
		TiempoEsperaEnvio: "120",
		EstadoEnvio:       record.EstadoEnvioCorrecto,
		RespuestaLinea: []record.RespuestaLinea{
			respuestaLineaPara(entry1, record.TipoOperacionAlta, record.EstadoRegistroCorrecto),
			respuestaLineaPara(entry2, record.TipoOperacionAlta, record.EstadoRegistroCorrecto),
			respuestaLineaPara(entry3, record.TipoOperacionAlta, record.EstadoRegistroCorrecto),
		},
	}

	envio, err := engine.Remitir(context.Background(), tenant)

	if err != nil {
		t.Fatalf("Error remitting: %v", err)
	}

	if envio == nil {
		t.Fatalf("Expected envio to be non-nil")
	}

	if envio.CSV != "CSV-DE-PRUEBA" {
		t.Errorf("Expected CSV to be 'CSV-DE-PRUEBA', got '%s'", envio.CSV)
	}

	if envio.TiempoEspera != 120*time.Second {
		t.Errorf("Expected TiempoEspera to be '120s', got '%s'", envio.TiempoEspera)
	}

	if len(envio.Lineas) != 3 {
		t.Fatalf("Expected 3 lineas, got %d", len(envio.Lineas))
	}

	if envio.Lineas[0].Secuencia != 1 || envio.Lineas[1].Secuencia != 2 || envio.Lineas[2].Secuencia != 3 {
		t.Fatalf("Expected lineas to have Secuencia 1, 2, 3, got %d, %d, %d", envio.Lineas[0].Secuencia, envio.Lineas[1].Secuencia, envio.Lineas[2].Secuencia)
	}

	if tf.llamadas != 1 {
		t.Errorf("Expected 1 call to transporteFalso, got %d", tf.llamadas)
	}

	if tf.ultimoTenant != tenant {
		t.Errorf("Expected ultimoTenant to be %v, got %v", tenant, tf.ultimoTenant)
	}

	if len(tf.ultimoLote.RegistroFactura) != 3 {
		t.Fatalf("Expected 3 RegistroFactura, got %d", len(tf.ultimoLote.RegistroFactura))
	}

	if tf.ultimoLote.Cabecera.ObligadoEmision.NIF != tenant.NIF {
		t.Errorf("Expected NIF in ultimoLote to be %s, got %s", tenant.NIF, tf.ultimoLote.Cabecera.ObligadoEmision.NIF)

	}

	if tf.ultimoLote.Cabecera.ObligadoEmision.NombreRazon != "EMPRESA DE PRUEBAS SL" {
		t.Errorf("Expected NombreRazon in ultimoLote to be 'EMPRESA DE PRUEBAS SL', got %s", tf.ultimoLote.Cabecera.ObligadoEmision.NombreRazon)
	}

	_, err = engine.Remitir(context.Background(), tenant)

	if !errors.Is(err, verifactu.ErrSinPendientes) {
		t.Fatalf("Expected error %v, got %v", verifactu.ErrSinPendientes, err)
	}

	if tf.llamadas != 1 {
		t.Errorf("Expected 1 call to transporteFalso, got %d", tf.llamadas)
	}

}
