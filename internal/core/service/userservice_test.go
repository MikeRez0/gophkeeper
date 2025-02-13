package service_test

import (
	"context"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/auth"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port/mock"
	"github.com/MikeRez0/gophkeeper/internal/core/service"
	"github.com/MikeRez0/gophkeeper/internal/core/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type prepareMocks func(repo *mock.MockRepository)

func TestService_UserRegister(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userRegisterTest struct {
		name      string
		user      domain.User
		mock      prepareMocks
		expError  error
		expResult *domain.User
	}

	hashedPass, _ := utils.HashPassword("test")
	user := domain.User{
		Login:    "test",
		Password: hashedPass,
		ID:       1,
	}

	tests := []userRegisterTest{
		{
			name: "Register good",
			user: domain.User{Login: user.Login, Password: "test"},
			mock: func(repo *mock.MockRepository) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(nil, domain.ErrDataNotFound)
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&user, nil)
			},
			expError:  nil,
			expResult: &user,
		},
		{
			name: "Register already exists",
			user: domain.User{Login: user.Login, Password: "test"},
			mock: func(repo *mock.MockRepository) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(&user, nil)
			},
			expError:  domain.ErrConflictingData,
			expResult: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := mock.NewMockRepository(mockCtrl)
			ts := mock.NewMockTokenService(mockCtrl)
			test.mock(repo)

			s, err := service.NewUserService(repo, ts, logger)
			assert.NoError(t, err)

			result, err := s.RegisterUser(context.Background(), &test.user)

			assert.Equal(t, test.expResult, result)
			assert.Equal(t, test.expError, err)
		})
	}
}

func TestService_UserLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger, _ := zap.NewProduction()

	type userLoginTest struct {
		name     string
		user     domain.User
		mock     prepareMocks
		expError error
	}

	hashedPass, _ := utils.HashPassword("test")
	user := domain.User{
		Login:    "test",
		Password: hashedPass,
		ID:       1,
	}

	tests := []userLoginTest{
		{
			name: "Login good",
			user: domain.User{Login: user.Login, Password: "test", ID: 1},
			mock: func(repo *mock.MockRepository) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(&user, nil)
			},
			expError: nil,
		},
		{
			name: "Password bad",
			user: domain.User{Login: user.Login, Password: "hacker"},
			mock: func(repo *mock.MockRepository) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), user.Login).Return(&user, nil)
			},
			expError: domain.ErrInvalidCredentials,
		},
		{
			name: "Login bad",
			user: domain.User{Login: "hacker", Password: "test"},
			mock: func(repo *mock.MockRepository) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), "hacker").Return(nil, domain.ErrDataNotFound)
			},
			expError: domain.ErrInvalidCredentials,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := mock.NewMockRepository(mockCtrl)
			ts, err := auth.New()
			assert.NoError(t, err)

			test.mock(repo)

			s, err := service.NewUserService(repo, ts, logger)
			assert.NoError(t, err)

			token, err := s.LoginUser(context.Background(), test.user.Login, test.user.Password)
			assert.Equal(t, test.expError, err)

			if token != "" {
				payload, err := ts.VerifyToken(token)
				assert.NoError(t, err)
				assert.Equal(t, payload.UserID, test.user.ID)
			}
		})
	}
}
