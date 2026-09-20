package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

// EnvioDe implements [verifactu.Store].
func (s *Store) EnvioDe(ctx context.Context, t verifactu.Tenant, secuencia uint64) (*verifactu.Envio, error) {
	consulta := `
	SELECT e.id, e.instante, e.csv, e.nif_presentador, e.timestamp_presentacion, e.estado_envio, e.tiempo_espera_segundos
	 FROM envios e
	 JOIN lineas l ON l.envio_id = e.id
	WHERE e.tenant_nif = ? AND e.tenant_sistema = ? AND l.secuencia = ?
	ORDER BY e.id DESC
	LIMIT 1
`

	row := s.db.QueryRowContext(ctx, consulta, t.NIF, t.IDSistemaInformatico, secuencia)

	return s.unEnvio(ctx, row)
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
	consulta := `
	SELECT id, instante, csv, nif_presentador, timestamp_presentacion, estado_envio, tiempo_espera_segundos
  	FROM envios
 	WHERE tenant_nif = ? AND tenant_sistema = ?
 	ORDER BY id DESC
 	LIMIT 1
	`

	row := s.db.QueryRowContext(ctx, consulta, t.NIF, t.IDSistemaInformatico)

	return s.unEnvio(ctx, row)

}

func (s *Store) unEnvio(ctx context.Context, row *sql.Row) (*verifactu.Envio, error) {
	var id, segundos int64
	var instante, timestampPresentacion, estado, csv, nifPresentador string

	if err := row.Scan(&id, &instante, &csv, &nifPresentador, &timestampPresentacion, &estado, &segundos); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, verifactu.ErrNoEncontrado
		}
		return nil, err
	}

	tiempoEspera := time.Duration(segundos) * time.Second
	instanteParseado, err := time.Parse(time.RFC3339, instante)
	if err != nil {
		return nil, err
	}

	timestampPresentacionParseado, err := time.Parse(time.RFC3339, timestampPresentacion)
	if err != nil {
		return nil, err
	}

	lineas, err := s.lineasDe(ctx, id)
	if err != nil {
		return nil, err
	}

	return &verifactu.Envio{
		Instante:              instanteParseado,
		CSV:                   csv,
		NIFPresentador:        nifPresentador,
		TimestampPresentacion: timestampPresentacionParseado,
		EstadoEnvio:           record.EstadoEnvio(estado),
		Lineas:                lineas,
		TiempoEspera:          tiempoEspera,
	}, nil
}

func (s *Store) lineasDe(ctx context.Context, envioID int64) ([]verifactu.LineaEnvio, error) {
	consulta := `
	SELECT secuencia, operacion, factura_nif, factura_num_serie, factura_fecha, estado, codigo_error, descripcion, duplicado
	 FROM lineas
	WHERE envio_id = ?
	ORDER BY secuencia ASC
	`

	rows, err := s.db.QueryContext(ctx, consulta, envioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lineas []verifactu.LineaEnvio
	for rows.Next() {
		var linea verifactu.LineaEnvio

		var operacion, estado, fecha string
		var duplicado sql.NullString

		if err := rows.Scan(&linea.Secuencia, &operacion, &linea.IDFactura.NIF, &linea.IDFactura.NumSerie, &fecha, &estado, &linea.CodigoError, &linea.Descripcion, &duplicado); err != nil {
			return nil, err
		}

		fechaParseada, err := record.ParseFecha(fecha)
		if err != nil {
			return nil, err
		}

		linea.Operacion = verifactu.Operacion(operacion)
		linea.Estado = record.EstadoRegistro(estado)

		linea.IDFactura.Fecha = fechaParseada
		if duplicado.Valid {
			var r record.RegistroDuplicado
			if err := json.Unmarshal([]byte(duplicado.String), &r); err != nil {
				return nil, err
			}
			linea.Duplicado = &r
		}

		lineas = append(lineas, linea)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lineas, nil
}
