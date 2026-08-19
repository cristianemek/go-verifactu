package verifactu

import (
	"context"
	"errors"
	"sync"
	"time"

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
// to compare two IDFactura values, use the Equal method.
type IDFactura struct {
	NIF      string
	NumSerie string
	Fecha    record.Fecha
}

// Equal returns true if two IDFactura values refer to the same invoice
func (id IDFactura) Equal(other IDFactura) bool {
	return id.NIF == other.NIF && id.NumSerie == other.NumSerie && id.Fecha.Format() == other.Fecha.Format()
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

// The id of an Entry is the IDFactura and the Operacion. Two entries with the same IDFactura and Operacion refer to the same invoice.
func (e *Entry) Matches(idFactura IDFactura, op Operacion) bool {
	return e.IDFactura.Equal(idFactura) && e.Operacion == op
}

// Config holds the configuration for a new Engine.
type Config struct {
	Store Store
	// Now is optional. If nil, time.Now is used. It is useful for testing.
	Now func() time.Time
}

// Engine owns the chain state. It serializes operations per tenant so two
// concurrent altas cannot fork a chain.
type Engine struct {
	store       Store
	now         func() time.Time
	mu          sync.Mutex
	tenantLocks map[Tenant]*sync.Mutex
}

func New(cfg Config) (*Engine, error) {
	if cfg.Store == nil {
		return nil, ErrStoreRequerido
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Engine{
		store:       cfg.Store,
		now:         cfg.Now,
		tenantLocks: make(map[Tenant]*sync.Mutex),
	}, nil
}

// lock serializes operations on a tenant. It returns the release function
func (e *Engine) lock(t Tenant) func() {
	e.mu.Lock()

	mu, ok := e.tenantLocks[t]

	if !ok {
		mu = &sync.Mutex{}
		e.tenantLocks[t] = mu
	}

	e.mu.Unlock()

	mu.Lock()

	return mu.Unlock

}

// Engine overwrites the following fields of the record you pass to it, without prompting:
// - Encadenamiento, FechaHoraHusoGenRegistro, Huella, TipoHuella, IDVersion
func (e *Engine) Alta(ctx context.Context, t Tenant, r record.RegistroAlta) (*Entry, error) {
	release := e.lock(t)
	defer release()

	id := IDFactura{
		NIF:      r.IDFactura.IDEmisorFactura,
		NumSerie: r.IDFactura.NumSerieFactura,
		Fecha:    r.IDFactura.FechaExpedicionFactura,
	}

	got, err := e.store.Buscar(ctx, t, id, OperacionAlta)

	if err != nil && !errors.Is(err, ErrNoEncontrado) {
		return nil, err
	}

	if err == nil {
		return got, nil
	}

	var secuencia uint64
	var encadenamiento record.Encadenamiento

	last, err := e.store.Ultimo(ctx, t)

	switch {
	case errors.Is(err, ErrNoEncontrado):
		secuencia = 1
		encadenamiento = record.NewEncadenamientoPrimerRegistro()

	case err != nil:
		return nil, err

	default:
		secuencia = last.Secuencia + 1
		encadenamiento = record.NewEncadenamientoRegistroAnterior(last.IDFactura.NIF, last.IDFactura.NumSerie, last.IDFactura.Fecha, last.Huella)
	}

	r.Encadenamiento = encadenamiento
	r.FechaHoraHusoGenRegistro = record.FechaHora(e.now().Truncate(time.Second))

	registroAlta, err := record.NewRegistroAlta(r)

	if err != nil {
		return nil, err
	}

	entry := Entry{
		Operacion: OperacionAlta,
		Alta:      &registroAlta,
		Secuencia: secuencia,
		Huella:    registroAlta.Huella,
		IDFactura: id,
		Anulacion: nil,
	}

	err = e.store.Anexar(ctx, t, &entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}
