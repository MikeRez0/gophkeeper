package port

import (
	"context"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

//go:generate mockgen -source=repository.go -destination=mock/repository.go -package=mock
type UserRepository interface {
	// User
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}

type KeyChainRepository interface {
	KeyChainInsert(context.Context, *domain.KCData) (*domain.KCData, error)

	KeyChainItemUpsert(context.Context, *domain.KCItemData) (*domain.KCItemData, error)
	KeyChainItemSelect(context.Context, domain.KeyChainID, domain.KeyChainItemID) (*domain.KCItemData, error)

	KeyChainGetItemsSince(ctx context.Context, keyChainID domain.KeyChainID,
		since time.Time) (*[]domain.KCItemData, error)
	KeyChainGetItemsFind(ctx context.Context, keyChainID domain.KeyChainID,
		label string, metas []domain.KeyChainItemMeta) (*[]domain.KCItemData, error)
}
