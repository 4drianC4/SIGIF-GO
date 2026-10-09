package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed sql/*.sql
var files embed.FS

const noVersion = -1

type Migration struct {
	Version uint
	Name    string
	Applied bool
}

type migrator struct {
	*migrate.Migrate
	source source.Driver
}

func open(dsn string) (*migrator, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	src, err := iofs.New(files, "sql")
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		_ = driver.Close()
		return nil, err
	}
	return &migrator{Migrate: m, source: src}, nil
}

func (m *migrator) current() (version int, dirty bool, err error) {
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return noVersion, false, nil
	}
	return int(v), dirty, err
}

func (m *migrator) previous(version int) (int, error) {
	prev, err := m.source.Prev(uint(version))
	if errors.Is(err, os.ErrNotExist) {
		return noVersion, nil
	}
	return int(prev), err
}

func (m *migrator) ensureClean() error {
	version, dirty, err := m.current()
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("database is marked dirty at version %d by an interrupted migration: check the schema and run `make migrate-force VERSION=<last applied version>`", version)
	}
	return nil
}

func (m *migrator) restoreAfterFailure(cause error, lastApplied func(failed int) (int, error)) error {
	failed, dirty, err := m.current()
	if err != nil {
		return errors.Join(cause, err)
	}
	if !dirty {
		return cause
	}
	version, err := lastApplied(failed)
	if err != nil {
		return errors.Join(cause, err)
	}
	if err := m.Force(version); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func Up(dsn string) error {
	m, err := open(dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.ensureClean(); err != nil {
		return err
	}
	if err := m.Migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return m.restoreAfterFailure(err, m.previous)
	}
	return nil
}

func Down(dsn string) error {
	m, err := open(dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.ensureClean(); err != nil {
		return err
	}
	before, _, err := m.current()
	if err != nil {
		return err
	}
	if before == noVersion {
		return nil
	}
	if err := m.Steps(-1); err != nil {
		return m.restoreAfterFailure(err, func(int) (int, error) { return before, nil })
	}
	return nil
}

func Force(dsn string, version int) error {
	m, err := open(dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Migrate.Force(version)
}

func Status(dsn string) ([]Migration, error) {
	m, err := open(dsn)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	current, _, err := m.current()
	if err != nil {
		return nil, err
	}

	var list []Migration
	version, err := m.source.First()
	for err == nil {
		reader, name, readErr := m.source.ReadUp(version)
		if readErr != nil {
			return nil, readErr
		}
		_ = reader.Close()
		list = append(list, Migration{Version: version, Name: name, Applied: int(version) <= current})
		version, err = m.source.Next(version)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return list, nil
}
