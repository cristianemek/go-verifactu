package verifactu

import (
	"context"

	"github.com/cristianemek/go-verifactu/record"
)

var _ Transport = (*transporteFalso)(nil)

type transporteFalso struct {
	respuesta record.RespuestaRegFactuSistemaFacturacion
	err       error

	llamadas     int
	ultimoLote   record.RegFactuSistemaFacturacion
	ultimoTenant Tenant
}

func (tf *transporteFalso) Remitir(ctx context.Context, t Tenant, lote record.RegFactuSistemaFacturacion) (record.RespuestaRegFactuSistemaFacturacion, error) {
	tf.llamadas++
	tf.ultimoLote = lote
	tf.ultimoTenant = t
	return tf.respuesta, tf.err
}
