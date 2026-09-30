package main

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

type transporteFalso struct {
	llamado  bool
	llamadas atomic.Int32
	estado   record.EstadoRegistro
	codigo   string
	err      error
}

func (tf *transporteFalso) Remitir(ctx context.Context, t verifactu.Tenant, lote record.RegFactuSistemaFacturacion) (record.RespuestaRegFactuSistemaFacturacion, error) {
	tf.llamado = true
	tf.llamadas.Add(1)

	if tf.err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, tf.err
	}

	estado := tf.estado
	if estado == "" {
		estado = record.EstadoRegistroCorrecto
	}

	respuesta := record.RespuestaRegFactuSistemaFacturacion{CSV: "A-1", EstadoEnvio: record.EstadoEnvioCorrecto}

	if estado != record.EstadoRegistroCorrecto {
		respuesta.EstadoEnvio = record.EstadoEnvioIncorrecto
	}

	for _, r := range lote.RegistroFactura {
		if r.RegistroAlta == nil {
			continue
		}

		respuesta.RespuestaLinea = append(respuesta.RespuestaLinea, record.RespuestaLinea{
			IDFactura:           r.RegistroAlta.IDFactura,
			Operacion:           record.OperacionRespuesta{TipoOperacion: record.TipoOperacionAlta},
			EstadoRegistro:      estado,
			CodigoErrorRegistro: tf.codigo,
		})
	}

	return respuesta, nil
}

func TestTransportePorTenantEnruta(t *testing.T) {
	uno := &transporteFalso{}
	dos := &transporteFalso{}

	tp := &transportePorTenant{
		transportes: map[string]verifactu.Transport{
			"89890001K": uno,
			"B12345674": dos,
		},
	}

	_, err := tp.Remitir(context.Background(), verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}, record.RegFactuSistemaFacturacion{})
	if err != nil {
		t.Fatalf("Remitir() = %v, want nil", err)
	}

	if !uno.llamado {
		t.Errorf("Remitir() did not call the correct transport for tenant 89890001K")
	}

	if dos.llamado {
		t.Errorf("Remitir() called the wrong transport for tenant B12345674")
	}
}

func TestTransportePorTenantSinCertificado(t *testing.T) {
	uno := &transporteFalso{}

	tp := &transportePorTenant{
		transportes: map[string]verifactu.Transport{
			"00000000X": uno,
		},
	}

	_, err := tp.Remitir(context.Background(), verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}, record.RegFactuSistemaFacturacion{})
	if err == nil {
		t.Fatalf("Remitir() = nil, want error for missing transport")
	}
}
