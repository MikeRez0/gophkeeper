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

type IKeychainRepository interface {
	// KeychainInsert(context.Context, *domain.KCData) (*domain.KCData, error)
	// KeychainUpdate(context.Context, *domain.KCData) (*domain.KCData, error)
	KeychainUpsert(context.Context, *domain.KCData) (*domain.KCData, error)
	KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error)
	KeychainGet(ctx context.Context, keychainID domain.KeychainID) (*domain.KCData, error)

	KeychainItemUpsert(context.Context, *domain.KCItemData) (*domain.KCItemData, bool, error)
	KeychainItemSelect(context.Context, domain.KeychainID, domain.KeychainItemID) (*domain.KCItemData, error)

	KeychainGetItemsSince(ctx context.Context, keyChainID domain.KeychainID,
		sinceClient time.Time, sinceServer time.Time) ([]*domain.KCItemData, error)
}
