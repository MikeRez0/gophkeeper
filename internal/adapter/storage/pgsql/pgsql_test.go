package pgsql_test

import (
	"context"
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
		db, err := pgsql.NewDBStorage(context.Background(), &config.Database{DSN: dbtest.DSN})
		if err != nil {
			return nil, err
		}
		err = db.RunMigrations()
		if err != nil {
			return nil, err
		}
		pgDB = db
	}

	return pgDB, nil
}

func TestPgRepository_User(t *testing.T) {
	log := logger.NewLogger(&config.App{LogLevel: "debug"})

	db, err := getPgDB()
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repoU, err := pgsql.NewUserRepository(db)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repotest.UserRepositoryTest(t, repoU)

	repoK, err := pgsql.NewKeychainPgRepository(db, log)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	repotest.KeychainRepositoryTest(t, repoK)
}
