package service_test

import (
	"context"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/repository"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/MikeRez0/gophkeeper/internal/core/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func getKeychainDeps(log *zap.Logger) (port.IKeychainRepository, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}

	urepo, err := repository.NewUserRepository(db)
	if err != nil {
		return nil, err
	}
	_, err = urepo.CreateUser(context.Background(), &domain.User{Login: "dummy"})
	if err != nil {
		return nil, err
	}

	repo, err := repository.NewKeychainPgRepository(db, log)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func TestKeychainService_Sync(t *testing.T) {
	l := logger.NewLogger(&config.App{LogLevel: "debug"})

	repo, err := getKeychainDeps(l)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	s, err := service.NewKeychainService(repo, l)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	userID := domain.UserID(1)
	k := &domain.KCData{
		ID:      domain.KeychainID(uuid.New()),
		OwnerID: userID,
	}

	ctx := context.Background()

	t.Run("Keychain save/get", func(t *testing.T) {
		k, err = s.KeychainCreate(ctx, userID, k)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		testK, err := s.KeychainGet(ctx, userID, k.ID)
		assert.NoError(t, err)
		assert.Equal(t, testK.ID, k.ID)

		_, err = s.KeychainGet(ctx, userID, domain.KeychainID(uuid.New()))
		assert.ErrorIs(t, err, domain.ErrDataNotFound)

	})
}
