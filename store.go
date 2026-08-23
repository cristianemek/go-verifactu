package verifactu

import "context"

// Store is an append-only ledger. There is no update or delete: a billing
// record is never modified.
// An adapter may demand additional restrictions on the tenant and return ErrTenantInvalido.
type Store interface {
	Ultimo(ctx context.Context, t Tenant) (*Entry, error)
	Anexar(ctx context.Context, t Tenant, e *Entry) error
	Buscar(ctx context.Context, t Tenant, idFactura IDFactura, op Operacion) (*Entry, error)
}
