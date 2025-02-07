package service

import (
	"context"
	"errors"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"go.uber.org/zap"
)

type KeyChainService struct {
	repo   port.IKeyChainRepository
	logger *zap.Logger
}

func NewKeychainService(repo port.IKeyChainRepository, logger *zap.Logger) (*KeyChainService, error) {
	return &KeyChainService{
		repo:   repo,
		logger: logger,
	}, nil
}

func (s *KeyChainService) KeychainCreate(ctx context.Context,
	userID domain.UserID,
	keychain *domain.KCData) (*domain.KCData, error) {
	if userID != keychain.OwnerID {
		return nil, domain.ErrBadRequest
	}

	oldK, err := s.repo.KeyChainGet(ctx, keychain.ID)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		return nil, err
	}

	if oldK == nil {
		kc, err := s.repo.KeyChainInsert(ctx, keychain)
		if err != nil {
			return nil, err
		}
		return kc, nil
	} else {
		if s.checkAuthority(ctx, userID, oldK.ID) {
			kc, err := s.repo.KeyChainUpdate(ctx, keychain)
			if err != nil {
				return nil, err
			}
			return kc, nil
		} else {
			return nil, domain.ErrForbidden
		}
	}
}

func (s *KeyChainService) KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error) {
	list, err := s.repo.KeyChainList(ctx, user)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *KeyChainService) KeychainGet(ctx context.Context, user domain.UserID,
	keychainID domain.KeychainID) (*domain.KCData, error) {
	list, err := s.repo.KeyChainList(ctx, user)
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

func (s *KeyChainService) KeychainSaveItem(ctx context.Context, user domain.UserID,
	item *domain.KCItemData) (*domain.KCItemData, bool, error) {
	if !s.checkAuthority(ctx, user, item.KeyChainID) {
		return nil, false, domain.ErrForbidden
	}
	item.ServerTime = time.Now()
	item, updated, err := s.repo.KeyChainItemUpsert(ctx, item)
	if err != nil {
		return nil, updated, err
	}
	return item, updated, nil
}
func (s *KeyChainService) KeychainGetItem(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
	id domain.KeychainItemID) (*domain.KCItemData, error) {
	if !s.checkAuthority(ctx, user, keychainID) {
		return nil, domain.ErrForbidden
	}

	item, err := s.repo.KeyChainItemSelect(ctx, keychainID, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *KeyChainService) KeychainGetItemsSince(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
	since time.Time) ([]*domain.KCItemData, error) {
	if !s.checkAuthority(ctx, user, keychainID) {
		return nil, domain.ErrForbidden
	}

	items, err := s.repo.KeyChainGetItemsSince(ctx, keychainID, since)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *KeyChainService) checkAuthority(ctx context.Context, userID domain.UserID, keychainID domain.KeychainID) bool {
	if k, err := s.repo.KeyChainGet(ctx, keychainID); err == nil {
		return k.OwnerID == userID
	}
	return false
}

func (s *KeyChainService) Sync(ctx context.Context, user domain.UserID,
	keychainID domain.KeychainID, fromTime time.Time, items []*domain.KCItemData) ([]*domain.KCItemData, error) {
	for _, i := range items {
		if keychainID != i.KeyChainID {
			return nil, domain.ErrBadRequest
		}
	}

	for _, i := range items {
		_, _, err := s.repo.KeyChainItemUpsert(ctx, i)
		if err != nil {
			return nil, err
		}
	}

	return s.KeychainGetItemsSince(ctx, user, keychainID, fromTime)
}
