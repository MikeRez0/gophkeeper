package service

import (
	"context"
	"errors"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"go.uber.org/zap"
)

// KeychainDataService is service controlling keychain data store.
// It has server and local modes.
// Server mode used for cloud storage and uses server date-time for synchronisation.
// Local mode used for local client storage and uses client date-time for synchronisation.
type KeychainDataService struct {
	repo      port.IKeychainRepository
	logger    *zap.Logger
	localMode bool
}

// NewKeychainDataService creates new KeychainDataService.
func NewKeychainDataService(repo port.IKeychainRepository, logger *zap.Logger) (*KeychainDataService, error) {
	return &KeychainDataService{
		repo:   repo,
		logger: logger,
	}, nil
}

// KeychainSave - create or update keychain by user.
func (s *KeychainDataService) KeychainSave(ctx context.Context,
	userID domain.UserID,
	keychain *domain.KCData) (*domain.KCData, error) {
	if !s.localMode && userID != keychain.OwnerID {
		return nil, domain.ErrBadRequest
	}

	oldK, err := s.repo.KeychainGet(ctx, keychain.ID)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		return nil, err //nolint:wrapcheck // it's ok
	}

	if oldK != nil && !s.checkAuthority(ctx, userID, oldK.ID) {
		return nil, domain.ErrForbidden
	}
	kc, err := s.repo.KeychainUpsert(ctx, keychain)
	if err != nil {
		return nil, err //nolint:wrapcheck // it's ok
	}
	return kc, nil
}

// KeychainList - list keychains by user.
func (s *KeychainDataService) KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error) {
	list, err := s.repo.KeychainList(ctx, user)
	if err != nil {
		return nil, err //nolint:wrapcheck // it's ok
	}
	return list, nil
}

// KeychainGet - get keychain header by user.
func (s *KeychainDataService) KeychainGet(ctx context.Context, user domain.UserID,
	keychainID domain.KeychainID) (*domain.KCData, error) {
	list, err := s.repo.KeychainList(ctx, user)
	if err != nil {
		return nil, err //nolint:wrapcheck // it's ok
	}
	for _, k := range list {
		if k.ID == keychainID {
			return k, nil
		}
	}
	return nil, domain.ErrDataNotFound
}

// KeychainSaveItem - create or update keychain item by user.
func (s *KeychainDataService) KeychainSaveItem(ctx context.Context, user domain.UserID,
	item *domain.KCItemData) (*domain.KCItemData, bool, error) {
	if !s.checkAuthority(ctx, user, item.KeyChainID) {
		return nil, false, domain.ErrForbidden
	}
	item.ServerTime = time.Now().UTC()
	item, updated, err := s.repo.KeychainItemUpsert(ctx, item)
	if err != nil {
		return nil, updated, err //nolint:wrapcheck // it's ok
	}
	return item, updated, nil
}

// KeychainGetItem - read keychain item by user.
func (s *KeychainDataService) KeychainGetItem(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
	id domain.KeychainItemID) (*domain.KCItemData, error) {
	if !s.checkAuthority(ctx, user, keychainID) {
		return nil, domain.ErrForbidden
	}

	item, err := s.repo.KeychainItemSelect(ctx, keychainID, id)
	if err != nil {
		return nil, err //nolint:wrapcheck // it's ok
	}
	return item, nil
}

// KeychainGetItemsSince - get keychain items updated after [since] time.
func (s *KeychainDataService) KeychainGetItemsSince(ctx context.Context,
	user domain.UserID, keychainID domain.KeychainID,
	since time.Time) ([]*domain.KCItemData, error) {
	if !s.checkAuthority(ctx, user, keychainID) {
		return nil, domain.ErrForbidden
	}

	var (
		sinceClient time.Time
		sinceServer time.Time
	)
	if s.localMode {
		sinceClient = since
	} else {
		sinceServer = since
	}

	items, err := s.repo.KeychainGetItemsSince(ctx, keychainID, sinceClient, sinceServer)
	if err != nil {
		return nil, err //nolint:wrapcheck // it's ok
	}
	return items, nil
}

// checkAuthority checks user authorization for keychain.
// Only owner authorizated to access keychain.
func (s *KeychainDataService) checkAuthority(ctx context.Context,
	userID domain.UserID, keychainID domain.KeychainID) bool {
	if s.localMode {
		return true
	}
	if k, err := s.repo.KeychainGet(ctx, keychainID); err == nil {
		return k.OwnerID == userID
	}
	return false
}

// Sync - sync items: save incoming items and return changed from [fromTime] time.
func (s *KeychainDataService) Sync(ctx context.Context, user domain.UserID,
	keychainID domain.KeychainID, fromTime time.Time, items []*domain.KCItemData) ([]*domain.KCItemData, error) {
	for _, i := range items {
		if keychainID != i.KeyChainID {
			return nil, domain.ErrBadRequest
		}
	}

	for _, i := range items {
		i.ServerTime = time.Now().UTC()
		_, _, err := s.repo.KeychainItemUpsert(ctx, i)
		if err != nil {
			return nil, err //nolint:wrapcheck // it's ok
		}
	}

	return s.KeychainGetItemsSince(ctx, user, keychainID, fromTime)
}

// SetLocalMode sets mode for service.
func (s *KeychainDataService) SetLocalMode(localMode bool) {
	s.localMode = localMode
}
