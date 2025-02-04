package service

import (
	"context"
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

func (s *KeyChainService) KeychainCreate(ctx context.Context, keychain *domain.KCData) (*domain.KCData, error) {
	kc, err := s.repo.KeyChainInsert(ctx, keychain)
	if err != nil {
		return nil, err
	}
	return kc, nil
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
	item *domain.KCItemData) (*domain.KCItemData, error) {
	item, err := s.repo.KeyChainItemUpsert(ctx, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
func (s *KeyChainService) KeychainGetItem(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
	id domain.KeychainItemID) (*domain.KCItemData, error) {
	item, err := s.repo.KeyChainItemSelect(ctx, keychainID, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *KeyChainService) KeychainGetItemsSince(ctx context.Context, user domain.UserID, keychainID domain.KeychainID,
	since time.Time) ([]*domain.KCItemData, error) {
	items, err := s.repo.KeyChainGetItemsSince(ctx, keychainID, since)
	if err != nil {
		return nil, err
	}
	return items, nil
}
