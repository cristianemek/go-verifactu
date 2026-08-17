package verifactu

import (
	"github.com/cristianemek/go-verifactu/record"
)

// Operacion tells which kind of record an Entry holds.
type Operacion string

const (
	OperacionAlta      Operacion = "alta"
	OperacionAnulacion Operacion = "anulacion"
)

// Tenant identifies one chain of records. Each record's fingerprint links to
// the previous one in the same chain, so operations on a tenant must be
// serialized.
type Tenant struct {
	NIF                  string
	IDSistemaInformatico string
}

// IDFactura identifies one unique invoice.Two calls with the same IDFactura should refer to the same invoice.
type IDFactura struct {
	NIF      string
	NumSerie string
	Fecha    record.Fecha
}

// Entry is one record in the ledger, holding either an Alta or an Anulacion.
// Once written it is never modified.
type Entry struct {
	Operacion Operacion
	Alta      *record.RegistroAlta
	Anulacion *record.RegistroAnulacion
	// Secuencia is the position in the chain, starting at 1. It lets the Store
	// spot a forked chain.
	Secuencia uint64
	// Huella is copied from the record, so lookups skip the pointer.
	Huella    string
	IDFactura IDFactura
}
