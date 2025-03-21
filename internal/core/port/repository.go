package port

import (
	"context"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

//go:generate mockgen -source=repository.go -destination=mock/repository.go -package=mock

// IUserRepository - interface for access data with users.
type IUserRepository interface {
	// CreateUser creates new user.
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// GetUserByLogin finds user.
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}

// IKeychainRepository - interface for access data with keychain and items.
type IKeychainRepository interface {
	// KeychainUpsert creates or updates keychain header.
	KeychainUpsert(context.Context, *domain.KCData) (*domain.KCData, error)
	// KeychainList returns list of keychain header.
	KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error)
	// KeychainGet returns keychain header.
	KeychainGet(ctx context.Context, keychainID domain.KeychainID) (*domain.KCData, error)

	// KeychainItemUpsert creates or updates keychain item.
	// Before upaded controls that item is newer than saved (by client datetime).
	KeychainItemUpsert(context.Context, *domain.KCItemData) (*domain.KCItemData, bool, error)
	// KeychainItemSelect returns keychain item.
	KeychainItemSelect(context.Context, domain.KeychainID, domain.KeychainItemID) (*domain.KCItemData, error)

	// KeychainGetItemsSince returns keychain items updated after client and/or server datetime.
	KeychainGetItemsSince(ctx context.Context, keyChainID domain.KeychainID,
		sinceClient time.Time, sinceServer time.Time) ([]*domain.KCItemData, error)
}
