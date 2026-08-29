package verifactu_test

import (
	"context"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

type transporteFalso struct {
	respuesta record.RespuestaRegFactuSistemaFacturacion
	err       error

	llamadas     int
	ultimoLote   record.RegFactuSistemaFacturacion
	ultimoTenant verifactu.Tenant
}

func (tf *transporteFalso) Remitir(ctx context.Context, t verifactu.Tenant, lote record.RegFactuSistemaFacturacion) (record.RespuestaRegFactuSistemaFacturacion, error) {
	tf.llamadas++
	tf.ultimoLote = lote
	tf.ultimoTenant = t
	return tf.respuesta, tf.err
}

var _ verifactu.Transport = (*transporteFalso)(nil)
