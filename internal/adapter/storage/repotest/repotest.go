package repotest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func UserRepositoryTest(t *testing.T, repo port.IUserRepository) {
	t.Helper()
	type useTest struct {
		run  func() error
		name string
	}

	ctx := context.Background()
	tests := []useTest{
		{
			name: "Create user",
			run: func() error {
				_, err := repo.CreateUser(ctx, &domain.User{Login: "Test", Password: "Test"})
				assert.NoError(t, err)
				return err
			},
		},
		{
			name: "Get user",
			run: func() error {
				user, err := repo.GetUserByLogin(ctx, "Test")
				assert.NoError(t, err)
				assert.Equal(t, "Test", user.Login)

				return nil
			},
		},
		{
			name: "Create user fail with violation",
			run: func() error {
				_, err := repo.CreateUser(ctx, &domain.User{Login: "Test", Password: "Test"})
				assert.ErrorIs(t, err, domain.ErrConflictingData)
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			assert.NoError(t, err, "test failed with error")
		})
	}
}

func KeychainRepositoryTest(t *testing.T, repo port.IKeychainRepository) {
	t.Helper()
	type useTest struct {
		run  func() error
		name string
	}

	ctx := context.Background()

	userID := domain.UserID(1)
	keychainID := domain.KeychainID(uuid.New())
	ki1 := domain.KCItemData{
		KeyChainID: keychainID,
		ID:         domain.KeychainItemID(uuid.New()),
		Label:      "Test label",
		ItemType:   domain.KCItemTypeString,
		Value:      []byte("test value"),
		Key:        []byte("test key"),
	}

	assertItemEq := func(t *testing.T, exp, act *domain.KCItemData) error {
		t.Helper()
		if assert.Equal(t, exp.ID, act.ID) &&
			assert.Equal(t, exp.KeyChainID, act.KeyChainID) &&
			assert.Equal(t, exp.Label, act.Label) &&
			assert.Equal(t, exp.Key, act.Key) &&
			assert.Equal(t, exp.Value, act.Value) &&
			assert.Equal(t, exp.ItemType, act.ItemType) &&
			assert.Equal(t, exp.ClientTime.UTC(), act.ClientTime.UTC()) &&
			assert.Equal(t, exp.ServerTime.UTC(), act.ServerTime.UTC()) {
			return nil
		} else {
			return errors.New("items doesn't equal")
		}
	}

	tests := []useTest{
		{
			name: "Create keychain",
			run: func() error {
				_, err := repo.KeychainUpsert(ctx, &domain.KCData{Name: "TestChain", OwnerID: userID, ID: keychainID})
				assert.NoError(t, err)
				if err != nil {
					return err
				}
				_, err = repo.KeychainUpsert(ctx, &domain.KCData{Name: "DummyChain",
					OwnerID: userID, ID: domain.KeychainID(uuid.New())})
				assert.NoError(t, err)
				return err
			},
		},
		{
			name: "Get keychain",
			run: func() error {
				k, err := repo.KeychainGet(ctx, keychainID)
				assert.NoError(t, err)
				if !assert.NotNil(t, k) {
					return errors.New("keychain is nil")
				}
				assert.Equal(t, "TestChain", k.Name)
				assert.Equal(t, keychainID, k.ID)
				assert.Equal(t, userID, k.OwnerID)

				return nil
			},
		},
		{
			name: "Get keychain list",
			run: func() error {
				l, err := repo.KeychainList(ctx, userID)
				assert.NoError(t, err)
				if !assert.NotNil(t, l) {
					return errors.New("keychain list is nil")
				}
				assert.Equal(t, 2, len(l))
				for _, k := range l {
					if k.Name == "TestChain" {
						assert.Equal(t, keychainID, k.ID)
						assert.Equal(t, userID, k.OwnerID)
						return nil
					}
				}

				return errors.New("keychain not found in list")
			},
		},
		{
			name: "Get keychain not found",
			run: func() error {
				k, err := repo.KeychainGet(ctx, domain.KeychainID(uuid.New()))
				assert.ErrorIs(t, err, domain.ErrDataNotFound)
				assert.Nil(t, k)

				return nil
			},
		},
		{
			name: "Add keychain item 1",
			run: func() error {
				ki1.ClientTime = time.Now().Truncate(1 * time.Millisecond)
				ki1.ServerTime = time.Now().Truncate(1 * time.Millisecond)
				k, updated, err := repo.KeychainItemUpsert(ctx, &ki1)
				assert.NoError(t, err)
				if !assert.NotNil(t, k) {
					return errors.New("keychain item is nil")
				}
				assert.True(t, updated)

				return assertItemEq(t, &ki1, k)
			},
		},
		{
			name: "Update keychain item 1",
			run: func() error {
				ki1.ClientTime = ki1.ClientTime.Add(1 * time.Minute)
				ki1.ServerTime = time.Now()
				ki1.Label += " (updated)"
				k, updated, err := repo.KeychainItemUpsert(ctx, &ki1)
				assert.NoError(t, err)
				if !assert.NotNil(t, k) {
					return errors.New("keychain item is nil")
				}
				assert.True(t, updated)

				return assertItemEq(t, &ki1, k)
			},
		},
		{
			name: "Get updated keychain item 1",
			run: func() error {
				k, err := repo.KeychainItemSelect(ctx, ki1.KeyChainID, ki1.ID)
				assert.NoError(t, err)
				if !assert.NotNil(t, k) {
					return errors.New("keychain item is nil")
				}

				return assertItemEq(t, &ki1, k)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			assert.NoError(t, err, "test failed with error")
		})
	}
}
