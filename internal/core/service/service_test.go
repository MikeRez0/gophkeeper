package service_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/pgsql"
	"github.com/MikeRez0/gophkeeper/internal/test/db"
)

var dbtest *db.TestDBInstance

func setup() {
	var err error
	dbtest, err = db.NewTestDBInstance(50000)
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
	database *pgsql.DB
	dbmutex  sync.Mutex
)

func getDB() (*pgsql.DB, error) {
	dbmutex.Lock()
	defer dbmutex.Unlock()

	if database == nil {
		d, err := pgsql.NewDBStorage(context.Background(), &config.Database{DSN: dbtest.DSN})
		if err != nil {
			return nil, fmt.Errorf("create database error: %w", err)
		}
		err = d.RunMigrations()
		if err != nil {
			return nil, fmt.Errorf("migration error: %w", err)
		}
		database = d
	}

	return database, nil
}
