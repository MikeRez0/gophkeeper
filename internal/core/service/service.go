package service

import (
	"context"
	"errors"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/MikeRez0/gophkeeper/internal/core/utils"
	"go.uber.org/zap"
)

type Service struct {
	repo         port.Repository
	tokenService port.TokenService
	logger       *zap.Logger
}

func NewService(repo port.Repository, tokenService port.TokenService, logger *zap.Logger) (*Service, error) {
	return &Service{
		repo:         repo,
		tokenService: tokenService,
		logger:       logger,
	}, nil
}

func (s *Service) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	exUser, err := s.repo.GetUserByLogin(ctx, user.Login)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		s.logger.Error("Get user", zap.Error(err))
		return nil, domain.ErrInternal
	}

	if exUser != nil {
		return nil, domain.ErrConflictingData
	}

	// Hash password
	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		s.logger.Error("Hash password", zap.Error(err))
		return nil, domain.ErrInternal
	}
	user.Password = hashed

	newUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error("Create user", zap.Error(err))
		return nil, domain.ErrInternal
	}

	return newUser, nil
}

func (s *Service) LoginUser(ctx context.Context, login string, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrDataNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", domain.ErrInternal
	}

	err = utils.ComparePassword(password, user.Password)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token, err := s.tokenService.CreateToken(user)
	if err != nil {
		s.logger.Error("Create token", zap.Error(err))
		return "", domain.ErrTokenCreation
	}

	return token, nil
}
