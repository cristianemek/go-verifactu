package verifactu

import (
	"context"

	"github.com/cristianemek/go-verifactu/record"
)

// Transport is the interface for sending requests to the Verifactu service.
type Transport interface {
	Remitir(ctx context.Context, t Tenant, lote record.RegFactuSistemaFacturacion) (record.RespuestaRegFactuSistemaFacturacion, error)
}
