package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/cristianemek/go-verifactu"
	_ "modernc.org/sqlite"
)

//go:embed migration/*.sql
var migrationFS embed.FS

func migrar(db *sql.DB) error {

	migrations, err := fs.Glob(migrationFS, "migration/*.sql")
	if err != nil {
		return err
	}

	var version int
	err = db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return err
	}

	if version > len(migrations) {
		return fmt.Errorf("sqlite: schema version %d is newer than this binary knows (%d)", version, len(migrations))
	}

	for i := version; i < len(migrations); i++ {
		migrationFile := migrations[i]
		err = aplicarMigracion(db, migrationFile, i+1)
		if err != nil {
			return err
		}

	}

	return nil
}

type Store struct {
	db *sql.DB
}

func New(ruta string) (s *Store, err error) {
	db, err := sql.Open("sqlite", ruta)

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to ping database: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;")
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to set database options: %w", err)
	}

	err = migrar(db)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to migrate database: %w", err)
	}

	return &Store{db}, nil
}

func aplicarMigracion(db *sql.DB, migrationFile string, version int) error {
	sqlBytes, err := migrationFS.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("sqlite: failed to execute migration %s: %w", migrationFile, err)
	}

	_, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	if err != nil {
		return fmt.Errorf("sqlite: failed to update schema version to %d: %w", version, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("sqlite: failed to commit migration %s: %w", migrationFile, err)
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

var _ verifactu.Store = (*Store)(nil)
