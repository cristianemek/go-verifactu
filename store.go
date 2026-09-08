package verifactu

import "context"

// Store is an append-only ledger. There is no update or delete: a billing
// record is never modified.
// An adapter may demand additional restrictions on the tenant and return ErrTenantInvalido.
type Store interface {
	// Ultimo returns the last entry, or ErrNoEncontrado if the chain is empty.
	Ultimo(ctx context.Context, t Tenant) (*Entry, error)

	// Anexar appends an entry. It must reject atomically one whose Secuencia is not
	// the next, with ErrConflictoDeSecuencia: that is what stops two writers from
	// forking the chain. The Engine retries on it.
	Anexar(ctx context.Context, t Tenant, e *Entry) error

	// Buscar returns that entry, or ErrNoEncontrado.
	Buscar(ctx context.Context, t Tenant, idFactura IDFactura, op Operacion) (*Entry, error)

	// Pendientes returns what is not settled yet, in chain order. A limite of zero
	// or less means no limit, and an empty result is not an error.
	Pendientes(ctx context.Context, t Tenant, limite int) ([]*Entry, error)

	// AnexarEnvio records one submission and marks the lines it settled.
	AnexarEnvio(ctx context.Context, t Tenant, envio *Envio) error

	// UltimoEnvio returns the last submission, or ErrNoEncontrado if there is none.
	UltimoEnvio(ctx context.Context, t Tenant) (*Envio, error)
}
