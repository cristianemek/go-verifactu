package verifactu

import "errors"

var (
	ErrNoEncontrado = errors.New("verifactu: entry not found")
	ErrDuplicado    = errors.New("verifactu: entry already exists")
	// ErrCadenaBifurcada means someone tried to append out of order. Never
	// recover from this silently: a forked chain cannot be fixed.
	ErrCadenaBifurcada   = errors.New("verifactu: chain forked")
	ErrStoreRequerido    = errors.New("verifactu: store is required")
	ErrTenantInvalido    = errors.New("verifactu: tenant is invalid")
	ErrOpcionNoAplicable = errors.New("verifactu: option not applicable for this operation")
)
