package verifactu

import "errors"

var (
	ErrNoEncontrado = errors.New("verifactu: entry not found")
	ErrDuplicado    = errors.New("verifactu: entry already exists")
	// ErrCadenaBifurcada means someone tried to append out of order. Never
	// recover from this silently: a forked chain cannot be fixed.
	ErrCadenaBifurcada      = errors.New("verifactu: chain forked")
	ErrStoreRequerido       = errors.New("verifactu: store is required")
	ErrTenantInvalido       = errors.New("verifactu: tenant is invalid")
	ErrOpcionNoAplicable    = errors.New("verifactu: option not applicable for this operation")
	ErrFaultServidor        = errors.New("verifactu: server fault, retry again")
	ErrFaultCliente         = errors.New("verifactu: client fault, bad request, check the data")
	ErrTransportRequerido   = errors.New("verifactu: transport is required")
	ErrObligadoDesconocido  = errors.New("verifactu: no name for the taxpayer: the batch has no Alta and no ConObligado was given")
	ErrSinPendientes        = errors.New("verifactu: no pending entries to send")
	ErrRespuestaDescuadrada = errors.New("verifactu: the AEAT answer does not line up with the batch sent")
	ErrEsperaActiva         = errors.New("verifactu: waiting, cannot send more entries yet")
)
