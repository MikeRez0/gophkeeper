package service_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/auth"
	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/repository"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/MikeRez0/gophkeeper/internal/core/service"
	"github.com/MikeRez0/gophkeeper/internal/e2etest/testdb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var dbtest *testdb.TestDBInstance

func setup() {
	var err error
	dbtest, err = testdb.NewTestDBInstance()
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

func getDeps() (port.Repository, port.TokenService, error) {
	db, err := storage.NewDBStorage(context.Background(), &config.Database{DSN: dbtest.DSN})
	if err != nil {
		return nil, nil, err
	}
	err = db.RunMigrations()
	if err != nil {
		return nil, nil, err
	}
	repo, err := repository.NewRepository(db)
	if err != nil {
		return nil, nil, err
	}
	ts, err := auth.New()
	if err != nil {
		return nil, nil, err
	}

	return repo, ts, nil
}

func TestServiceDB_UserRegister(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userRegisterTest struct {
		name      string
		user      domain.User
		expError  error
		expResult *domain.User
	}

	tests := []userRegisterTest{
		{
			name:      "Register good",
			user:      domain.User{Login: "test", Password: "test"},
			expError:  nil,
			expResult: &domain.User{Login: "test"},
		},
		{
			name:      "Register good",
			user:      domain.User{Login: "test2", Password: "test"},
			expError:  nil,
			expResult: &domain.User{Login: "test2"},
		},
		{
			name:      "Register already exists",
			user:      domain.User{Login: "test", Password: "test"},
			expError:  domain.ErrConflictingData,
			expResult: nil,
		},
	}

	repo, ts, err := getDeps()
	assert.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := service.NewService(repo, ts, logger)
			assert.NoError(t, err)

			result, err := s.RegisterUser(context.Background(), &test.user)

			if test.expResult != nil {
				assert.Equal(t, test.expResult.Login, result.Login)
			}
			assert.Equal(t, test.expError, err)
		})
	}
}

func TestServiceDB_UserLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userLoginTest struct {
		name         string
		registerUser bool
		user         domain.User
		expError     error
	}

	tests := []userLoginTest{
		{
			registerUser: true,
			name:         "Login good",
			user:         domain.User{Login: "test", Password: "test"},

			expError: nil,
		},
		{
			name:     "Password bad",
			user:     domain.User{Login: "test", Password: "hacker"},
			expError: domain.ErrInvalidCredentials,
		},
		{
			name:     "Login bad",
			user:     domain.User{Login: "hacker", Password: "test"},
			expError: domain.ErrInvalidCredentials,
		},
	}

	repo, ts, err := getDeps()
	assert.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := service.NewService(repo, ts, logger)
			assert.NoError(t, err)

			if test.registerUser {
				s.RegisterUser(context.Background(), &test.user)
			}

			token, err := s.LoginUser(context.Background(), test.user.Login, test.user.Password)
			assert.Equal(t, test.expError, err)

			if token != "" {
				_, err := ts.VerifyToken(token)
				assert.NoError(t, err)
			}
		})
	}
}
