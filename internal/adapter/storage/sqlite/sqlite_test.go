package sqlite_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"

	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/repotest"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/sqlite"
	"github.com/stretchr/testify/assert"
)

var (
	sqliteDB      *sqlite.DB
	sqliteDBmutex sync.Mutex
)

const cSqliteFilename = "test.db"

func setup() {}
func shutdown() {
	if _, err := os.Stat(cSqliteFilename); err == nil {
		err = os.Remove(cSqliteFilename)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func getSqliteDB() (*sqlite.DB, error) {
	sqliteDBmutex.Lock()
	defer sqliteDBmutex.Unlock()

	if sqliteDB == nil {
		database, err := sqlite.NewDBStorage(context.Background(), &config.Database{DSN: cSqliteFilename, Driver: "sqlite3"})
		if err != nil {
			return nil, fmt.Errorf("create database error: %w", err)
		}
		err = database.RunMigrations()
		if err != nil {
			return nil, fmt.Errorf("migrate error: %w", err)
		}
		sqliteDB = database
	}

	return sqliteDB, nil
}

func TestSqliteRepository(t *testing.T) {
	l := logger.NewLogger(&config.App{LogLevel: "debug"})
	database, err := getSqliteDB()
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repoU, err := sqlite.NewUserRepository(database, l)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repotest.UserRepositoryTest(t, repoU)

	repoK, err := sqlite.NewKeychainSqliteRepository(database, l)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repotest.KeychainRepositoryTest(t, repoK)
}
