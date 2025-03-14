package port

import (
	"context"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

// IUserService - interface for service that manages users.
type IUserService interface {
	// RegisterUser - Register new user
	RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// LoginUser - Authenticate user
	LoginUser(ctx context.Context, login string, password string) (string, error)
}

// IKeychainDataService - interface for keychain store service. Manage keychains and items.
type IKeychainDataService interface {
	// KeychainSave - create or update keychain by user.
	KeychainSave(ctx context.Context, user domain.UserID, keychain *domain.KCData) (*domain.KCData, error)
	// KeychainList - list keychains by user.
	KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error)
	// KeychainGet - get keychain header by user.
	KeychainGet(ctx context.Context, user domain.UserID, keychainID domain.KeychainID) (*domain.KCData, error)

	// KeychainSaveItem - create or update keychain item by user.
	KeychainSaveItem(ctx context.Context, user domain.UserID,
		item *domain.KCItemData) (*domain.KCItemData, bool, error)
	// KeychainGetItem - read keychain item by user.
	KeychainGetItem(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
		id domain.KeychainItemID) (*domain.KCItemData, error)

	// KeychainGetItemsSince - get keychain items updated after [since] time.
	KeychainGetItemsSince(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
		since time.Time) ([]*domain.KCItemData, error)

	// Sync - sync items: save incoming items and return changed from [fromTime] time.
	Sync(ctx context.Context, user domain.UserID,
		keychainID domain.KeychainID, fromTime time.Time, items []*domain.KCItemData) ([]*domain.KCItemData, error)
}
