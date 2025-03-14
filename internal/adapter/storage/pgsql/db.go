// Package pgsql contains repository implementation for PostgreSQL database.
package pgsql

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
)

// DB is database object with PgSQL engine.
type DB struct {
	*pgxpool.Pool
	QueryBuilder *squirrel.StatementBuilderType
	dsn          string
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

// NewDBStorage creates new DB object.
func NewDBStorage(ctx context.Context, conf *config.Database) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	dbs := DB{
		Pool:         pool,
		dsn:          conf.DSN,
		QueryBuilder: &psql,
	}

	return &dbs, nil
}

// RunMigrations runs migrations on database.
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
