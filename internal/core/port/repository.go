package port

import (
	"context"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

//go:generate mockgen -source=repository.go -destination=mock/repository.go -package=mock
type Repository interface {
	// User
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}
