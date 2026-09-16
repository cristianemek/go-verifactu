package main

import (
	"context"
	"fmt"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

type transportePorTenant struct {
	transportes map[string]verifactu.Transport
}

var _ verifactu.Transport = (*transportePorTenant)(nil)

func (tp *transportePorTenant) Remitir(ctx context.Context, t verifactu.Tenant, lote record.RegFactuSistemaFacturacion) (record.RespuestaRegFactuSistemaFacturacion, error) {
	transporte, ok := tp.transportes[t.NIF]
	if !ok {
		return record.RespuestaRegFactuSistemaFacturacion{}, fmt.Errorf("no certificate for tenant %s", t.NIF)
	}

	return transporte.Remitir(ctx, t, lote)
}
