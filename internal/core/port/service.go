package port

import (
	"context"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

type Service interface {
	RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error)
	LoginUser(ctx context.Context, login string, password string) (string, error)
}
