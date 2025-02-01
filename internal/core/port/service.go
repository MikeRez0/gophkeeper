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

type IKeyChainDataService interface {
	KeyChainCreate(ctx context.Context, owner domain.UserID) (*domain.KCData, error)

	KeyChainSaveItem(ctx context.Context, item *domain.KCItemData) (*domain.KCItemData, error)
	KeyChainGetItem(ctx context.Context, keyChainID domain.KeyChainID,
		id domain.KeyChainItemID) (*domain.KCItemData, error)

	KeyChainGetItemsSince(ctx context.Context, keyChainID domain.KeyChainID,
		since time.Time) (*[]domain.KCItemData, error)
}
