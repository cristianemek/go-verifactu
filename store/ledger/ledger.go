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

var _ verifactu.Store = (*Store)(nil)

// Store keeps one append-only JSONL file per tenant inside a directory, with a
// single entry per line. New rebuilds the in-memory index from disk, so a chain
// survives a restart.
//
// This adapter requires both Tenant fields to be non-empty and uppercase
// alphanumeric ASCII, because together they become the file name. Anything else
// returns ErrTenantInvalido, which is what keeps a crafted NIF from writing
// outside the directory.
//
// Only Anexar enforces that: Ultimo and Buscar answer from the index, so an
// invalid tenant gets ErrNoEncontrado from them rather than ErrTenantInvalido.
type Store struct {
	dir        string
	mu         sync.RWMutex
	cadenas    map[verifactu.Tenant][]*verifactu.Entry
	envios     map[verifactu.Tenant][]*verifactu.Envio
	liquidados map[verifactu.Tenant]map[uint64]bool
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
	envios := make(map[verifactu.Tenant][]*verifactu.Envio)
	liquidados := make(map[verifactu.Tenant]map[uint64]bool)

	for _, entry := range dirEntry {
		nombre := entry.Name()

		if entry.IsDir() {
			continue
		}
		if filepath.Ext(nombre) != ".jsonl" {
			continue
		}

		esEnvios := strings.HasSuffix(nombre, ".envios.jsonl")

		var base string

		if esEnvios {
			base = strings.TrimSuffix(nombre, ".envios.jsonl")
		} else {
			base = strings.TrimSuffix(nombre, ".jsonl")
		}

		tenant, err := tenantDesdeNombre(base)
		if err != nil {
			return nil, err
		}

		dirPath := filepath.Join(dir, nombre)

		if esEnvios {
			enviosEntries, err := cargarJSONL[verifactu.Envio](dirPath)
			if err != nil {
				return nil, fmt.Errorf("error reading file '%s': %w", dirPath, err)
			}
			envios[tenant] = enviosEntries

			liquidados[tenant] = make(map[uint64]bool)
			for _, envio := range enviosEntries {
				marcarLiquidadas(liquidados[tenant], envio)
			}

		} else {
			entries, err := cargarJSONL[verifactu.Entry](dirPath)
			if err != nil {
				return nil, fmt.Errorf("error reading file '%s': %w", dirPath, err)
			}
			cadenas[tenant] = entries
		}
	}

	return &Store{
		dir:        dir,
		cadenas:    cadenas,
		envios:     envios,
		liquidados: liquidados,
	}, nil
}

func tenantDesdeNombre(base string) (verifactu.Tenant, error) {
	parts := strings.SplitN(base, "-", 2)

	if len(parts) != 2 || !alfanumerico(parts[0]) || !alfanumerico(parts[1]) {
		return verifactu.Tenant{}, fmt.Errorf("%w: invalid tenant file name '%s'", verifactu.ErrTenantInvalido, base)
	}

	return verifactu.Tenant{
		NIF:                  parts[0],
		IDSistemaInformatico: parts[1],
	}, nil
}

func marcarLiquidadas(dest map[uint64]bool, envio *verifactu.Envio) {
	for _, linea := range envio.Lineas {
		if linea.Liquidada() {
			dest[linea.Secuencia] = true
		}
	}
}

func nombreBase(tenant verifactu.Tenant) (string, error) {

	if !alfanumerico(tenant.NIF) {
		return "", fmt.Errorf("%w: NIF '%s' must be uppercase alphanumeric", verifactu.ErrTenantInvalido, tenant.NIF)
	}

	if !alfanumerico(tenant.IDSistemaInformatico) {
		return "", fmt.Errorf("%w: IDsistemaInformatico '%s' must be uppercase alphanumeric", verifactu.ErrTenantInvalido, tenant.IDSistemaInformatico)
	}

	return tenant.NIF + "-" + tenant.IDSistemaInformatico, nil
}

func (s *Store) fichero(tenant verifactu.Tenant) (string, error) {

	base, err := nombreBase(tenant)
	if err != nil {
		return "", err
	}

	return filepath.Join(s.dir, base+".jsonl"), nil
}

func (s *Store) ficheroEnvios(t verifactu.Tenant) (string, error) {
	base, err := nombreBase(t)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.dir, base+".envios.jsonl"), nil
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

	if !e.Correccion {
		for _, c := range s.cadenas[t] {
			if c.Matches(e.IDFactura, e.Operacion) {
				return verifactu.ErrDuplicado
			}
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

// cargarCadena reads a JSONL file and returns a slice of Entry pointers. If the file ends with an incomplete line, it truncates the file to remove that line.
func cargarJSONL[T any](dir string) ([]*T, error) {
	file, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	size := fileInfo.Size()

	var bytesCount int64 = 0
	var truncarA int64 = -1

	var entries []*T
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var lineNumber int
	for scanner.Scan() {
		scannerBytes := scanner.Bytes()

		bytesCount += int64(len(scannerBytes)) + 1

		if bytesCount > size {
			truncarA = bytesCount - int64(len(scannerBytes)) - 1
			break
		}

		lineNumber++

		if len(scannerBytes) == 0 {
			continue
		}

		var entry T

		if err := json.Unmarshal(scannerBytes, &entry); err != nil {
			return nil, fmt.Errorf("error unmarshalling on line %d: %w", lineNumber, err)
		}
		entries = append(entries, &entry)

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if truncarA >= 0 {
		if err := os.Truncate(dir, truncarA); err != nil {
			return nil, fmt.Errorf("error truncating file '%s' to %d bytes: %w", dir, truncarA, err)
		}
	}

	return entries, nil
}

// AnexarEnvio implements [verifactu.Store].
func (s *Store) AnexarEnvio(ctx context.Context, t verifactu.Tenant, envio *verifactu.Envio) error {
	dir, err := s.ficheroEnvios(t)
	if err != nil {
		return err
	}

	s.mu.Lock()

	defer s.mu.Unlock()

	jsonData, err := json.Marshal(envio)
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

	s.envios[t] = append(s.envios[t], envio)

	if s.liquidados[t] == nil {
		s.liquidados[t] = make(map[uint64]bool)
	}
	marcarLiquidadas(s.liquidados[t], envio)

	return nil
}

// Pendientes implements [verifactu.Store].
func (s *Store) Pendientes(ctx context.Context, t verifactu.Tenant, limite int) ([]*verifactu.Entry, error) {
	s.mu.RLock()

	defer s.mu.RUnlock()

	cadena := s.cadenas[t]

	liquidados := s.liquidados[t]

	var pendientes []*verifactu.Entry

	for _, e := range cadena {
		if !liquidados[e.Secuencia] {
			pendientes = append(pendientes, e)

			if limite > 0 && len(pendientes) >= limite {
				break
			}
		}
	}

	return pendientes, nil
}

// UltimoEnvio implements [verifactu.Store].
func (s *Store) UltimoEnvio(ctx context.Context, t verifactu.Tenant) (*verifactu.Envio, error) {
	s.mu.RLock()

	defer s.mu.RUnlock()

	envios := s.envios[t]

	if len(envios) == 0 {
		return nil, verifactu.ErrNoEncontrado
	}

	return envios[len(envios)-1], nil
}
