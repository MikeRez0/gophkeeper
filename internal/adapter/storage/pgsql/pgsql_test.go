package pgsql_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/pgsql"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/repotest"
	"github.com/MikeRez0/gophkeeper/internal/test/db"
	"github.com/stretchr/testify/assert"
)

var dbtest *db.TestDBInstance

func setup() {
	var err error
	dbtest, err = db.NewTestDBInstance(50001)
	if err != nil {
		log.Fatal(err)
	}
}
func shutdown() {
	if dbtest != nil {
		dbtest.Down()
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

var (
	pgDB    *pgsql.DB
	dbmutex sync.Mutex
)

func getPgDB() (*pgsql.DB, error) {
	dbmutex.Lock()
	defer dbmutex.Unlock()

	if pgDB == nil {
		database, err := pgsql.NewDBStorage(context.Background(), &config.Database{DSN: dbtest.DSN})
		if err != nil {
			return nil, fmt.Errorf("create database error: %w", err)
		}
		err = database.RunMigrations()
		if err != nil {
			return nil, fmt.Errorf("migrate error: %w", err)
		}
		pgDB = database
	}

	return pgDB, nil
}

func TestPgRepository_User(t *testing.T) {
	l := logger.NewLogger(&config.App{LogLevel: "debug"})

	database, err := getPgDB()
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repoU, err := pgsql.NewUserRepository(database)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repotest.UserRepositoryTest(t, repoU)

	repoK, err := pgsql.NewKeychainPgRepository(database, l)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repotest.KeychainRepositoryTest(t, repoK)
}
