package sqlite

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cristianemek/go-verifactu"
)

// EnvioDe implements [verifactu.Store].
func (s *Store) EnvioDe(ctx context.Context, t verifactu.Tenant, secuencia uint64) (*verifactu.Envio, error) {
	panic("unimplemented")
}

// AnexarEnvio implements [verifactu.Store].
func (s *Store) AnexarEnvio(ctx context.Context, t verifactu.Tenant, envio *verifactu.Envio) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	consulta := `INSERT INTO envios 
	(tenant_nif, tenant_sistema,instante,csv,nif_presentador,timestamp_presentacion, estado_envio, tiempo_espera_segundos) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.ExecContext(ctx, consulta, t.NIF, t.IDSistemaInformatico, envio.Instante.Format(time.RFC3339), envio.CSV, envio.NIFPresentador, envio.TimestampPresentacion.Format(time.RFC3339), string(envio.EstadoEnvio), int64(envio.TiempoEspera.Seconds()))
	if err != nil {
		return err
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	insertLinea := `INSERT INTO lineas
            (envio_id, tenant_nif, tenant_sistema, secuencia, operacion,
             factura_nif, factura_num_serie, factura_fecha, estado,
             codigo_error, descripcion, duplicado, procesada)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.PrepareContext(ctx, insertLinea)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, linea := range envio.Lineas {

		procesada := 0
		if linea.Procesada() {
			procesada = 1
		}

		var duplicado any = nil

		if linea.Duplicado != nil {
			data, err := json.Marshal(linea.Duplicado)
			if err != nil {
				return err
			}
			duplicado = string(data)
		}

		_, err := stmt.ExecContext(ctx,
			lastInsertId,
			t.NIF,
			t.IDSistemaInformatico,
			linea.Secuencia,
			string(linea.Operacion),
			linea.IDFactura.NIF,
			linea.IDFactura.NumSerie,
			linea.IDFactura.Fecha.Format(),
			string(linea.Estado),
			linea.CodigoError,
			linea.Descripcion,
			duplicado,
			procesada,
		)
		if err != nil {
			return err
		}

	}

	return tx.Commit()
}

// UltimoEnvio implements [verifactu.Store].
func (s *Store) UltimoEnvio(ctx context.Context, t verifactu.Tenant) (*verifactu.Envio, error) {
	panic("unimplemented")
}
