-- 0001_inicial.sql

CREATE TABLE entradas (
  tenant_nif        TEXT    NOT NULL,
  tenant_sistema    TEXT    NOT NULL,
  secuencia         INTEGER NOT NULL CHECK (secuencia >= 1),
  operacion         TEXT    NOT NULL CHECK (operacion IN ('alta', 'anulacion')),
  factura_nif       TEXT    NOT NULL,
  factura_num_serie TEXT    NOT NULL,
  factura_fecha     TEXT    NOT NULL
                    CHECK (factura_fecha GLOB '[0-3][0-9]-[0-1][0-9]-[0-9][0-9][0-9][0-9]'),
  huella            TEXT    NOT NULL,
  correccion        INTEGER NOT NULL CHECK (correccion IN (0, 1)),
  entrada           TEXT    NOT NULL CHECK (json_valid(entrada)),
  creado_en         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  PRIMARY KEY (tenant_nif, tenant_sistema, secuencia)
) STRICT;

CREATE INDEX idx_entradas_factura ON entradas
  (tenant_nif, tenant_sistema, factura_nif, factura_num_serie, factura_fecha, operacion, secuencia);

CREATE TRIGGER entradas_no_update BEFORE UPDATE ON entradas
BEGIN
  SELECT RAISE(ABORT, 'entradas es inmutable: no se permite UPDATE');
END;

CREATE TRIGGER entradas_no_delete BEFORE DELETE ON entradas
BEGIN
  SELECT RAISE(ABORT, 'entradas es inmutable: no se permite DELETE');
END;


CREATE TABLE envios (
  id                     INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_nif             TEXT    NOT NULL,
  tenant_sistema         TEXT    NOT NULL,
  instante               TEXT    NOT NULL,
  csv                    TEXT    NOT NULL,
  nif_presentador        TEXT    NOT NULL,
  timestamp_presentacion TEXT    NOT NULL,
  estado_envio           TEXT    NOT NULL,
  tiempo_espera_segundos INTEGER NOT NULL,
  creado_en              TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
) STRICT;

CREATE INDEX idx_envios_ultimo ON envios (tenant_nif, tenant_sistema, id DESC);

CREATE TRIGGER envios_no_update BEFORE UPDATE ON envios
BEGIN
  SELECT RAISE(ABORT, 'envios es inmutable: no se permite UPDATE');
END;

CREATE TRIGGER envios_no_delete BEFORE DELETE ON envios
BEGIN
  SELECT RAISE(ABORT, 'envios es inmutable: no se permite DELETE');
END;


CREATE TABLE lineas (
  envio_id          INTEGER NOT NULL REFERENCES envios (id),
  tenant_nif        TEXT    NOT NULL,
  tenant_sistema    TEXT    NOT NULL,
  secuencia         INTEGER NOT NULL,
  operacion         TEXT    NOT NULL,
  factura_nif       TEXT    NOT NULL,
  factura_num_serie TEXT    NOT NULL,
  factura_fecha     TEXT    NOT NULL,
  estado            TEXT    NOT NULL,
  codigo_error      TEXT    NOT NULL,
  descripcion       TEXT    NOT NULL,
  duplicado         TEXT    CHECK (duplicado IS NULL OR json_valid(duplicado)),
  procesada         INTEGER NOT NULL CHECK (procesada IN (0, 1)),
  PRIMARY KEY (envio_id, secuencia)
) WITHOUT ROWID, STRICT;

CREATE INDEX idx_lineas_secuencia ON lineas (tenant_nif, tenant_sistema, secuencia);

CREATE TRIGGER lineas_mismo_tenant BEFORE INSERT ON lineas
WHEN NOT EXISTS (SELECT 1 FROM envios
                  WHERE id = NEW.envio_id
                    AND tenant_nif = NEW.tenant_nif
                    AND tenant_sistema = NEW.tenant_sistema)
BEGIN
  SELECT RAISE(ABORT, 'la linea no pertenece al tenant del envio');
END;
