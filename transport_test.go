package verifactu_test

import (
	"context"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

var _ verifactu.Transport = (*transporteFalso)(nil)

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
