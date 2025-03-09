package service_test

import (
	"context"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/auth"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/pgsql"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/MikeRez0/gophkeeper/internal/core/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func getUserDeps() (port.IUserRepository, port.TokenService, error) {
	db, err := getDB()
	if err != nil {
		return nil, nil, err
	}

	repo, err := pgsql.NewUserRepository(db)
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

	repo, ts, err := getUserDeps()
	assert.NoError(t, err)
	if err != nil {
		return
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := service.NewUserService(repo, ts, logger)
			assert.NoError(t, err)

			result, err := s.RegisterUser(context.Background(), &test.user)
			if test.expError == nil {
				assert.NoError(t, err)
			}

			if test.expResult != nil {
				assert.NotNil(t, result)
				if result == nil {
					return
				}
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
			user:         domain.User{Login: "testuser", Password: "test"},

			expError: nil,
		},
		{
			name:     "Password bad",
			user:     domain.User{Login: "testuser", Password: "hacker"},
			expError: domain.ErrInvalidCredentials,
		},
		{
			name:     "Login bad",
			user:     domain.User{Login: "hacker", Password: "test"},
			expError: domain.ErrInvalidCredentials,
		},
	}

	repo, ts, err := getUserDeps()
	assert.NoError(t, err)
	if err != nil {
		return
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := service.NewUserService(repo, ts, logger)
			assert.NoError(t, err)

			if test.registerUser {
				u := domain.User{Login: test.user.Login, Password: test.user.Password}
				_, err := s.RegisterUser(context.Background(), &u)
				assert.NoError(t, err)
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
