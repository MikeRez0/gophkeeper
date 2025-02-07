package port

import (
	"context"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

//go:generate mockgen -source=repository.go -destination=mock/repository.go -package=mock
type IUserRepository interface {
	// User
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}

type IKeyChainRepository interface {
	KeyChainInsert(context.Context, *domain.KCData) (*domain.KCData, error)
	KeyChainUpdate(context.Context, *domain.KCData) (*domain.KCData, error)
	KeyChainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error)
	KeyChainGet(ctx context.Context, keychainID domain.KeychainID) (*domain.KCData, error)

	KeyChainItemUpsert(context.Context, *domain.KCItemData) (*domain.KCItemData, bool, error)
	KeyChainItemSelect(context.Context, domain.KeychainID, domain.KeychainItemID) (*domain.KCItemData, error)

	KeyChainGetItemsSince(ctx context.Context, keyChainID domain.KeychainID,
		since time.Time) ([]*domain.KCItemData, error)
}
