package port

import "github.com/MikeRez0/gophkeeper/internal/core/domain"

type TokenPayload struct {
	UserID domain.UserID
}

//go:generate mockgen -source=auth.go -destination=mock/auth.go -package=mock
type TokenService interface {
	CreateToken(user *domain.User) (string, error)
	VerifyToken(token string) (*TokenPayload, error)
}
