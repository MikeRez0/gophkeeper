package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

func (a *ClientApp) FetchKeychainList(ctx context.Context) ([]*domain.KCData, error) {
	data, err := a.doRequest(ctx, "/api/keychain", http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("error requesting keychain list : %w", err)
	}

	keychainList := make([]*domain.KCData, 0)
	err = json.Unmarshal(data, &keychainList)
	if err != nil {
		return nil, fmt.Errorf("error parsing response %w", err)
	}

	return keychainList, nil
}

func (a *ClientApp) SyncKeychains(ctx context.Context) error {
	keychainList, err := a.Service.KeychainList(ctx, a.UserID)
	if err != nil {
		return fmt.Errorf("get keychain list error: %w", err)
	}
	serverList, err := a.FetchKeychainList(ctx)
	if err != nil {
		return fmt.Errorf("fetch keychain list error: %w", err)
	}

	for _, s := range serverList {
		if !slices.ContainsFunc(keychainList, func(k *domain.KCData) bool { return k.ID == s.ID }) {
			keychainList = append(keychainList, s)
		}
		_, err := a.Service.KeychainSave(ctx, a.UserID, s)
		if err != nil {
			return fmt.Errorf("keychain header save error: %w", err)
		}
	}

	// Save current time for next sync time offset
	syncTime := time.Now().UTC()

	for _, keychain := range keychainList {
		// Keychain header sync
		_, err := a.doRequest(ctx,
			"/api/keychain/"+keychain.ID.String(),
			http.MethodPost,
			keychain)
		if err != nil {
			return fmt.Errorf("error on keychain data sync: %w", err)
		}

		// Keychain items sync
		items, err := a.Service.KeychainGetItemsSince(ctx, a.UserID, keychain.ID, a.SyncTime)
		if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
			return fmt.Errorf("get items error: %w", err)
		}

		query := ""
		if !a.SyncTime.IsZero() {
			query = "?from_time=" + a.SyncTime.Format(time.RFC3339)
		}

		data, err := a.doRequest(ctx,
			"/api/keychain/"+keychain.ID.String()+"/sync"+query,
			http.MethodPost,
			items,
		)
		if err != nil {
			return fmt.Errorf("error on keychain items sync: %w", err)
		}

		clear(items)

		err = json.Unmarshal(data, &items)
		if err != nil {
			return fmt.Errorf("error unmarshalling items:%w", err)
		}

		_, err = a.Service.Sync(ctx, a.UserID, keychain.ID, time.Time{}, items)
		if err != nil {
			return fmt.Errorf("error saving to local storage: %w", err)
		}
	}
	// Save sync time for next iteration
	a.SyncTime = syncTime

	return nil
}
