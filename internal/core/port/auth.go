package port

import "github.com/MikeRez0/gophkeeper/internal/core/domain"

// TokenPayload - data stored in token.
type TokenPayload struct {
	UserID domain.UserID
}

//go:generate mockgen -source=auth.go -destination=mock/auth.go -package=mock

// TokenService - interface for token control service.
type TokenService interface {
	// CreateToken - creates new token for given user.
	CreateToken(user *domain.User) (string, error)
	// VerifyToken - validate token and return payload.
	VerifyToken(token string) (*TokenPayload, error)
}
