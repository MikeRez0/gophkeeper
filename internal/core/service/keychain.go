package service

import (
	"context"
	"errors"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"go.uber.org/zap"
)

type KeychainDataService struct {
	repo      port.IKeychainRepository
	logger    *zap.Logger
	localMode bool
}

func NewKeychainDataService(repo port.IKeychainRepository, logger *zap.Logger) (*KeychainDataService, error) {
	return &KeychainDataService{
		repo:   repo,
		logger: logger,
	}, nil
}

func (s *KeychainDataService) KeychainSave(ctx context.Context,
	userID domain.UserID,
	keychain *domain.KCData) (*domain.KCData, error) {
	if !s.localMode && userID != keychain.OwnerID {
		return nil, domain.ErrBadRequest
	}

	oldK, err := s.repo.KeychainGet(ctx, keychain.ID)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		return nil, err
	}

	if oldK != nil && !s.checkAuthority(ctx, userID, oldK.ID) {
		return nil, domain.ErrForbidden
	}
	kc, err := s.repo.KeychainUpsert(ctx, keychain)
	if err != nil {
		return nil, err
	}
	return kc, nil
}

func (s *KeychainDataService) KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error) {
	list, err := s.repo.KeychainList(ctx, user)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *KeychainDataService) KeychainGet(ctx context.Context, user domain.UserID,
	keychainID domain.KeychainID) (*domain.KCData, error) {
	list, err := s.repo.KeychainList(ctx, user)
	if err != nil {
		return nil, err
	}
	for _, k := range list {
		if k.ID == keychainID {
			return k, nil
		}
	}
	return nil, domain.ErrDataNotFound
}

func (s *KeychainDataService) KeychainSaveItem(ctx context.Context, user domain.UserID,
	item *domain.KCItemData) (*domain.KCItemData, bool, error) {
	if !s.checkAuthority(ctx, user, item.KeyChainID) {
		return nil, false, domain.ErrForbidden
	}
	item.ServerTime = time.Now().UTC()
	item, updated, err := s.repo.KeychainItemUpsert(ctx, item)
	if err != nil {
		return nil, updated, err
	}
	return item, updated, nil
}
func (s *KeychainDataService) KeychainGetItem(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
	id domain.KeychainItemID) (*domain.KCItemData, error) {
	if !s.checkAuthority(ctx, user, keychainID) {
		return nil, domain.ErrForbidden
	}

	item, err := s.repo.KeychainItemSelect(ctx, keychainID, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *KeychainDataService) KeychainGetItemsSince(ctx context.Context,
	user domain.UserID, keychainID domain.KeychainID,
	since time.Time) ([]*domain.KCItemData, error) {
	if !s.checkAuthority(ctx, user, keychainID) {
		return nil, domain.ErrForbidden
	}

	items, err := s.repo.KeychainGetItemsSince(ctx, keychainID, since)
	if err != nil {
		return nil, err
	}
	return items, nil
}

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
			return nil, err
		}
	}

	return s.KeychainGetItemsSince(ctx, user, keychainID, fromTime)
}

func (s *KeychainDataService) SetLocalMode(localMode bool) {
	s.localMode = localMode
}
