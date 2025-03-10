package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/pgsql"
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

	urepo, err := pgsql.NewUserRepository(db)
	if err != nil {
		return nil, fmt.Errorf("create user repo err: %w", err)
	}
	_, err = urepo.CreateUser(context.Background(), &domain.User{Login: "dummy"})
	if err != nil {
		return nil, fmt.Errorf("create user error: %w", err)
	}

	repo, err := pgsql.NewKeychainPgRepository(db, log)
	if err != nil {
		return nil, fmt.Errorf("create keychain repo err: %w", err)
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

	s, err := service.NewKeychainDataService(repo, l)
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
		k, err = s.KeychainSave(ctx, userID, k)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		testK, err := s.KeychainGet(ctx, userID, k.ID)
		assert.NoError(t, err)
		if !assert.NotNil(t, testK) {
			return
		}
		if testK != nil {
			assert.Equal(t, testK.ID, k.ID)
		}

		_, err = s.KeychainGet(ctx, userID, domain.KeychainID(uuid.New()))
		assert.ErrorIs(t, err, domain.ErrDataNotFound)
	})
}
