package sqlite

import (
	"context"
	"encoding/json"

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
