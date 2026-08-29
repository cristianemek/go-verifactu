package memory

import (
	"context"
	"sync"

	"github.com/cristianemek/go-verifactu"
)

var _ verifactu.Store = (*Store)(nil)

type Store struct {
	mu      sync.RWMutex
	cadenas map[verifactu.Tenant][]*verifactu.Entry
}

func New() *Store {

	return &Store{
		cadenas: make(map[verifactu.Tenant][]*verifactu.Entry),
	}
}

func (s *Store) Ultimo(ctx context.Context, t verifactu.Tenant) (*verifactu.Entry, error) {
	s.mu.RLock()

	defer s.mu.RUnlock()

	cadena := s.cadenas[t]

	if len(cadena) == 0 {
		return nil, verifactu.ErrNoEncontrado
	}

	return cadena[len(cadena)-1], nil
}

func (s *Store) Buscar(ctx context.Context, t verifactu.Tenant, id verifactu.IDFactura, op verifactu.Operacion) (*verifactu.Entry, error) {
	s.mu.RLock()

	defer s.mu.RUnlock()

	for _, e := range s.cadenas[t] {
		if e.Matches(id, op) {
			return e, nil
		}
	}

	return nil, verifactu.ErrNoEncontrado
}

func (s *Store) Anexar(ctx context.Context, t verifactu.Tenant, e *verifactu.Entry) error {
	s.mu.Lock()

	defer s.mu.Unlock()

	cadena := s.cadenas[t]

	if e.Secuencia != uint64(len(cadena)+1) {
		return verifactu.ErrCadenaBifurcada
	}

	if !e.Correccion {
		for _, c := range cadena {
			if c.Matches(e.IDFactura, e.Operacion) {
				return verifactu.ErrDuplicado
			}
		}
	}

	s.cadenas[t] = append(cadena, e)

	return nil
}

// AnexarEnvio implements [verifactu.Store].
func (s *Store) AnexarEnvio(ctx context.Context, t verifactu.Tenant, envio *verifactu.Envio) error {
	panic("unimplemented")
}

// Pendientes implements [verifactu.Store].
func (s *Store) Pendientes(ctx context.Context, t verifactu.Tenant, limite int) ([]*verifactu.Entry, error) {
	panic("unimplemented")
}

// UltimoEnvio implements [verifactu.Store].
func (s *Store) UltimoEnvio(ctx context.Context, t verifactu.Tenant) (*verifactu.Envio, error) {
	panic("unimplemented")
}
