package memory

import (
	"context"
	"sync"

	"github.com/cristianemek/go-verifactu"
)

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

	for _, c := range cadena {
		if c.Matches(e.IDFactura, e.Operacion) {
			return verifactu.ErrDuplicado
		}
	}

	s.cadenas[t] = append(cadena, e)

	return nil
}
