package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/cristianemek/go-verifactu"
)

func (s *Store) Anexar(ctx context.Context, t verifactu.Tenant, e *verifactu.Entry) error {

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if !e.Correccion {
		var exists bool
		consulta :=
			`SELECT EXISTS(
  		SELECT 1 FROM entradas
   		 WHERE tenant_nif = ? AND tenant_sistema = ?
    		AND factura_nif = ? AND factura_num_serie = ? AND factura_fecha = ?
    		AND operacion = ?
		)`

		err := tx.QueryRowContext(ctx, consulta, t.NIF, t.IDSistemaInformatico, e.IDFactura.NIF, e.IDFactura.NumSerie, e.IDFactura.Fecha.Format(), string(e.Operacion)).Scan(&exists)

		if err != nil {
			return err
		}

		if exists {
			return verifactu.ErrDuplicado
		}
	}

	consultaInsert := `INSERT INTO entradas
	(tenant_nif, tenant_sistema, secuencia, operacion, factura_nif, factura_num_serie, factura_fecha, huella, correccion, entrada)
	SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	WHERE ? = COALESCE((SELECT MAX(secuencia) FROM entradas WHERE tenant_nif = ? AND tenant_sistema = ?), 0) + 1
	`

	correccionInt := 0
	if e.Correccion {
		correccionInt = 1
	}

	row, err := tx.ExecContext(ctx, consultaInsert, t.NIF, t.IDSistemaInformatico, e.Secuencia, string(e.Operacion), e.IDFactura.NIF, e.IDFactura.NumSerie, e.IDFactura.Fecha.Format(), e.Huella, correccionInt, string(data), e.Secuencia, t.NIF, t.IDSistemaInformatico)

	if err != nil {
		return err
	}

	rowsAffected, err := row.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return verifactu.ErrConflictoDeSecuencia
	}

	return tx.Commit()
}

func (s *Store) Ultimo(ctx context.Context, t verifactu.Tenant) (*verifactu.Entry, error) {
	consulta := `SELECT entrada FROM entradas
			 WHERE tenant_nif = ? AND tenant_sistema = ?
			 ORDER BY secuencia DESC
			 LIMIT 1`

	row := s.db.QueryRowContext(ctx, consulta, t.NIF, t.IDSistemaInformatico)

	return unaEntrada(row)

}

func (s *Store) Buscar(ctx context.Context, t verifactu.Tenant, id verifactu.IDFactura, op verifactu.Operacion) (*verifactu.Entry, error) {
	consulta := `SELECT entrada FROM entradas
					WHERE tenant_nif = ? AND tenant_sistema = ?
					AND factura_nif = ? AND factura_num_serie = ? AND factura_fecha = ?
					AND operacion = ?
					ORDER BY secuencia ASC
					LIMIT 1`

	row := s.db.QueryRowContext(ctx, consulta, t.NIF, t.IDSistemaInformatico, id.NIF, id.NumSerie, id.Fecha.Format(), string(op))

	return unaEntrada(row)
}

// AnexarEnvio implements [verifactu.Store].
func (s *Store) AnexarEnvio(ctx context.Context, t verifactu.Tenant, envio *verifactu.Envio) error {
	panic("unimplemented")
}

// Cadena implements [verifactu.Store].
func (s *Store) Cadena(ctx context.Context, t verifactu.Tenant) ([]*verifactu.Entry, error) {
	consulta := `SELECT entrada FROM entradas
 WHERE tenant_nif = ? AND tenant_sistema = ?
 ORDER BY secuencia ASC
`
	rows, err := s.db.QueryContext(
		ctx,
		consulta,
		t.NIF,
		t.IDSistemaInformatico,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*verifactu.Entry

	for rows.Next() {
		var datos []byte
		err := rows.Scan(&datos)

		if err != nil {
			return nil, err
		}

		var e verifactu.Entry

		err = json.Unmarshal(datos, &e)

		if err != nil {
			return nil, err
		}

		entries = append(entries, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// EnvioDe implements [verifactu.Store].
func (s *Store) EnvioDe(ctx context.Context, t verifactu.Tenant, secuencia uint64) (*verifactu.Envio, error) {
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

func unaEntrada(row *sql.Row) (*verifactu.Entry, error) {
	var datos []byte

	err := row.Scan(&datos)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, verifactu.ErrNoEncontrado
	}

	if err != nil {
		return nil, err
	}

	var e verifactu.Entry

	err = json.Unmarshal(datos, &e)

	if err != nil {
		return nil, err
	}

	return &e, nil
}
