package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
)

type DB struct {
	*sql.DB
	QueryBuilder *squirrel.StatementBuilderType
	dsn          string
	driver       string
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func NewDBStorage(ctx context.Context, conf *config.Database) (*DB, error) {
	if conf.Driver == "sqlite3" {
		_, err := os.Stat(conf.DSN)
		if os.IsNotExist(err) {
			_, err = os.Create(conf.DSN)
			if err != nil {
				return nil, fmt.Errorf("failed to create db file: %w", err)
			}
		}
	}

	db, err := sql.Open(conf.Driver, conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	dsn := conf.DSN
	if conf.Driver == "sqlite3" {
		dsn = "sqlite3://" + dsn
	}

	dbs := DB{
		DB:           db,
		driver:       conf.Driver,
		dsn:          dsn,
		QueryBuilder: &psql,
	}

	return &dbs, nil
}

func (db *DB) RunMigrations() error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, db.dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}
