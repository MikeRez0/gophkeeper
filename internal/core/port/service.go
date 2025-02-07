package port

import (
	"context"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

type IUserService interface {
	RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error)
	LoginUser(ctx context.Context, login string, password string) (string, error)
}

type IKeychainDataService interface {
	KeychainCreate(ctx context.Context, user domain.UserID, keychain *domain.KCData) (*domain.KCData, error)
	KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error)
	KeychainGet(ctx context.Context, user domain.UserID, keychainID domain.KeychainID) (*domain.KCData, error)

	KeychainSaveItem(ctx context.Context, user domain.UserID,
		item *domain.KCItemData) (*domain.KCItemData, bool, error)
	KeychainGetItem(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
		id domain.KeychainItemID) (*domain.KCItemData, error)

	KeychainGetItemsSince(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
		since time.Time) ([]*domain.KCItemData, error)

	Sync(ctx context.Context, user domain.UserID,
		keychainID domain.KeychainID, fromTime time.Time, items []*domain.KCItemData) ([]*domain.KCItemData, error)
}
