package ledger

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cristianemek/go-verifactu"
)

// Store is a ledger-based implementation of the verifactu.Store interface.
type Store struct {
	dir     string
	mu      sync.RWMutex
	cadenas map[verifactu.Tenant][]*verifactu.Entry
}

func New(dir string) (*Store, error) {

	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return nil, err
	}

	dirEntry, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	cadenas := make(map[verifactu.Tenant][]*verifactu.Entry)

	for _, entry := range dirEntry {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		tenantName := strings.TrimSuffix(entry.Name(), ".jsonl")
		parts := strings.SplitN(tenantName, "-", 2)

		if len(parts) != 2 || !alfanumerico(parts[0]) || !alfanumerico(parts[1]) {
			return nil, fmt.Errorf("%w: invalid tenant file name '%s'", verifactu.ErrTenantInvalido, entry.Name())

		}

		tenant := verifactu.Tenant{
			NIF:                  parts[0],
			IDSistemaInformatico: parts[1],
		}

		dirPath := filepath.Join(dir, entry.Name())

		entries, err := leerEntradasDesdeArchivo(dirPath)
		if err != nil {
			return nil, fmt.Errorf("error reading file '%s': %w", dirPath, err)
		}
		cadenas[tenant] = entries
	}

	return &Store{dir: dir, cadenas: cadenas}, nil

}

func (s *Store) fichero(tenant verifactu.Tenant) (string, error) {

	if !alfanumerico(tenant.NIF) {
		return "", fmt.Errorf("%w: NIF '%s' must be uppercase alphanumeric", verifactu.ErrTenantInvalido, tenant.NIF)
	}

	if !alfanumerico(tenant.IDSistemaInformatico) {
		return "", fmt.Errorf("%w: IDsistemaInformatico '%s' must be uppercase alphanumeric", verifactu.ErrTenantInvalido, tenant.IDSistemaInformatico)
	}

	nombre := tenant.NIF + "-" + tenant.IDSistemaInformatico + ".jsonl"

	return filepath.Join(s.dir, nombre), nil
}

func alfanumerico(s string) bool {

	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
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
	dir, err := s.fichero(t)
	if err != nil {
		return err
	}

	s.mu.Lock()

	defer s.mu.Unlock()

	if e.Secuencia != uint64(len(s.cadenas[t])+1) {
		return verifactu.ErrCadenaBifurcada
	}

	for _, c := range s.cadenas[t] {
		if c.Matches(e.IDFactura, e.Operacion) {
			return verifactu.ErrDuplicado
		}
	}

	jsonData, err := json.Marshal(e)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(jsonData, '\n')); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	s.cadenas[t] = append(s.cadenas[t], e)

	return nil
}

func leerEntradasDesdeArchivo(dir string) ([]*verifactu.Entry, error) {
	file, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []*verifactu.Entry
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var lineNumber int
	for scanner.Scan() {

		if len(scanner.Bytes()) == 0 {
			continue
		}

		var entry verifactu.Entry

		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, fmt.Errorf("error unmarshalling entry on line %d: %w", lineNumber, err)
		}
		entries = append(entries, &entry)

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
