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
	// Correccion means the entry shares its key with an earlier one on purpose.
	Correccion bool
}

// The id of an Entry is the IDFactura and the Operacion. Two entries with the same IDFactura and Operacion refer to the same invoice.
func (e *Entry) Matches(idFactura IDFactura, op Operacion) bool {
	return e.IDFactura.Equal(idFactura) && e.Operacion == op
}

// Config holds the configuration for a new Engine.
type Config struct {
	Store Store
	// Transport is optional. If nil, the Engine will not send records to the Verifactu service.
	Transport Transport
	// Now is optional. If nil, time.Now is used. It is useful for testing.
	Now func() time.Time
}

// Engine owns the chain state. It serializes operations per tenant so two
// concurrent altas cannot fork a chain.
type Engine struct {
	store       Store
	transport   Transport
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
		transport:   cfg.Transport,
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

// Assumes that the caller holds the lock for the tenant. Return the next sequence number and the encadenamiento for the next record.
func (e *Engine) siguienteCadena(ctx context.Context, t Tenant) (uint64, record.Encadenamiento, error) {
	last, err := e.store.Ultimo(ctx, t)

	switch {
	case errors.Is(err, ErrNoEncontrado):
		return 1, record.NewEncadenamientoPrimerRegistro(), nil
	case err != nil:
		return 0, record.Encadenamiento{}, err

	default:
		return last.Secuencia + 1, record.NewEncadenamientoRegistroAnterior(last.IDFactura.NIF, last.IDFactura.NumSerie, last.IDFactura.Fecha, last.Huella), nil
	}

}

// Alta records an issued invoice: it assigns the sequence number, builds the
// chain link, fixes the generation timestamp and persists the entry.
//
// It is idempotent: calling it again with the same IDFactura returns the entry
// that already exists. ComoSubsanacion and TrasRechazo turn that off, so the
// call appends a correction sharing the same key.
//
// The Engine overwrites these fields of the record you pass: Encadenamiento,
// FechaHoraHusoGenRegistro, IDVersion, TipoHuella, Huella, and with options
// Subsanacion and RechazoPrevio. The record is taken by value.
func (e *Engine) Alta(ctx context.Context, t Tenant, r record.RegistroAlta, opciones ...OpcionRegistro) (*Entry, error) {
	opts := aplicarOpcionesRegistro(opciones...)

	esCorreccion := opts.esSubsanacion || opts.esTrasRechazo

	release := e.lock(t)
	defer release()

	id := IDFactura{
		NIF:      r.IDFactura.IDEmisorFactura,
		NumSerie: r.IDFactura.NumSerieFactura,
		Fecha:    r.IDFactura.FechaExpedicionFactura,
	}

	if !esCorreccion {
		got, err := e.store.Buscar(ctx, t, id, OperacionAlta)

		if err != nil && !errors.Is(err, ErrNoEncontrado) {
			return nil, err
		}

		if err == nil {
			return got, nil
		}

	}

	secuencia, encadenamiento, err := e.siguienteCadena(ctx, t)

	if err != nil {
		return nil, err
	}

	r.Encadenamiento = encadenamiento
	r.FechaHoraHusoGenRegistro = record.FechaHora(e.now().Truncate(time.Second))
	if esCorreccion {
		if opts.esTrasRechazo {
			r.RechazoPrevio = record.Ptr(record.RechazoPrevioNoExiste)
		}
		r.Subsanacion = record.Ptr(record.SiNoSi)
	}

	registroAlta, err := record.NewRegistroAlta(r)

	if err != nil {
		return nil, err
	}

	entry := Entry{
		Operacion:  OperacionAlta,
		Alta:       &registroAlta,
		Secuencia:  secuencia,
		Huella:     registroAlta.Huella,
		IDFactura:  id,
		Anulacion:  nil,
		Correccion: esCorreccion,
	}

	err = e.store.Anexar(ctx, t, &entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

// Anular records the cancellation of a previously issued invoice. Same algorithm
// as Alta: it consumes a sequence number and links to the previous entry.
//
// Cancellation records are rarely the right tool. To undo an invoice the usual
// path is a corrective invoice recorded with Alta, which keeps both documents in
// the chain.
//
// It is idempotent unless TrasRechazo is passed. ComoSubsanacion does not apply
// to a cancellation and returns ErrOpcionNoAplicable.
//
// The Engine overwrites these fields of the record you pass: Encadenamiento,
// FechaHoraHusoGenRegistro, IDVersion, TipoHuella, Huella, and with TrasRechazo
// RechazoPrevio. The record is taken by value.
func (e *Engine) Anular(ctx context.Context, t Tenant, r record.RegistroAnulacion, opciones ...OpcionRegistro) (*Entry, error) {

	opts := aplicarOpcionesRegistro(opciones...)

	if opts.esSubsanacion {
		return nil, ErrOpcionNoAplicable
	}

	esCorreccion := opts.esSubsanacion || opts.esTrasRechazo

	release := e.lock(t)
	defer release()

	id := IDFactura{
		NIF:      r.IDFactura.IDEmisorFacturaAnulada,
		NumSerie: r.IDFactura.NumSerieFacturaAnulada,
		Fecha:    r.IDFactura.FechaExpedicionFacturaAnulada,
	}

	if !esCorreccion {
		got, err := e.store.Buscar(ctx, t, id, OperacionAnulacion)

		if err != nil && !errors.Is(err, ErrNoEncontrado) {
			return nil, err
		}

		if err == nil {
			return got, nil
		}

	}

	secuencia, encadenamiento, err := e.siguienteCadena(ctx, t)

	if err != nil {
		return nil, err
	}

	r.Encadenamiento = encadenamiento
	r.FechaHoraHusoGenRegistro = record.FechaHora(e.now().Truncate(time.Second))
	if esCorreccion {
		if opts.esTrasRechazo {
			r.RechazoPrevio = record.Ptr(record.RechazoPrevioAnulacionSi)
		}
	}

	registroAnulacion, err := record.NewRegistroAnulacion(r)
	if err != nil {
		return nil, err
	}

	entry := Entry{
		Operacion:  OperacionAnulacion,
		Alta:       nil,
		Anulacion:  &registroAnulacion,
		Secuencia:  secuencia,
		Huella:     registroAnulacion.Huella,
		IDFactura:  id,
		Correccion: esCorreccion,
	}

	err = e.store.Anexar(ctx, t, &entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

// Estado returns the current state of a given invoice, or ErrNoEncontrado if it is not found.
func (e *Engine) Estado(ctx context.Context, t Tenant, idFactura IDFactura, op Operacion) (*Entry, error) {

	return e.store.Buscar(ctx, t, idFactura, op)
}
