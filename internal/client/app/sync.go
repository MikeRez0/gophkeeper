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

// FetchKeychainList trying to fetch keychain headers from server.
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

// SyncKeychains trying to sync keychain items with server.
//  1. Read local keychain list
//  2. Merge keychain list from server
//
// For every keychain:
//  1. read local storage, fetch all items updated since last synchronization time [client_ts >= syncDate]
//  2. push items to server
//  3. save  items returned from server (items that created or changed on server since last synchronization time ).
//
// Returns syncTaskResult with information about pulled, pushed items.
func (a *ClientApp) SyncKeychains(ctx context.Context) (*syncTaskResult, error) {
	keychainList, err := a.Service.KeychainList(ctx, a.UserID)
	if err != nil {
		return nil, fmt.Errorf("get keychain list error: %w", err)
	}
	serverList, err := a.FetchKeychainList(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch keychain list error: %w", err)
	}

	for _, s := range serverList {
		if !slices.ContainsFunc(keychainList, func(k *domain.KCData) bool { return k.ID == s.ID }) {
			keychainList = append(keychainList, s)
		}
		_, err := a.Service.KeychainSave(ctx, a.UserID, s)
		if err != nil {
			return nil, fmt.Errorf("keychain header save error: %w", err)
		}
	}

	syncResult := syncTaskResult{}

	// Save current time for next sync time offset
	syncTime := time.Now().UTC()

	for _, keychain := range keychainList {
		// Keychain header sync
		_, err := a.doRequest(ctx,
			"/api/keychain/"+keychain.ID.String(),
			http.MethodPost,
			keychain)
		if err != nil {
			return nil, fmt.Errorf("error on keychain data sync: %w", err)
		}

		// 1. read local storage, fetch all items updated since last synchronization time [client_ts >= syncDate]
		localItems, err := a.Service.KeychainGetItemsSince(ctx, a.UserID, keychain.ID, a.syncTime)
		if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
			return nil, fmt.Errorf("get items error: %w", err)
		}

		// 2. push items to server
		query := ""
		if !a.syncTime.IsZero() {
			query = "?from_time=" + a.syncTime.Format(time.RFC3339)
		}

		syncResult.Uploaded = len(localItems)

		data, err := a.doRequest(ctx,
			"/api/keychain/"+keychain.ID.String()+"/sync"+query,
			http.MethodPost,
			localItems,
		)
		if err != nil {
			return nil, fmt.Errorf("error on keychain items sync: %w", err)
		}

		// 3. save  items returned from server (items that created or changed on server since last synchronization time ).
		serverItems := make([]*domain.KCItemData, 0)

		err = json.Unmarshal(data, &serverItems)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling items:%w", err)
		}

		syncResult.Downloaded = len(serverItems)

		for _, i := range serverItems {
			if slices.ContainsFunc(localItems, func(v *domain.KCItemData) bool { return v.ID == i.ID }) {
				syncResult.UpdatedForNewerVersion++
			}
		}

		_, err = a.Service.Sync(ctx, a.UserID, keychain.ID, time.Time{}, serverItems)
		if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
			return nil, fmt.Errorf("error saving to local storage: %w", err)
		}
	}
	// Save sync time for next iteration
	a.syncTime = syncTime

	return &syncResult, nil
}

type syncTaskResult struct {
	Uploaded               int
	Downloaded             int
	UpdatedForNewerVersion int
}

type syncTaskStatus struct {
	Error          error
	Finished       bool
	SyncTaskResult syncTaskResult
}

var (
	ErrSyncActive = errors.New("can't start new sync proccess, it's already active")
)

// RunSync starts sync job.
// Returns new chanel, which will be informed about job status.
func (a *ClientApp) RunSync(ctx context.Context) (chan syncTaskStatus, error) {
	if a.syncActive.Load() {
		return nil, ErrSyncActive
	}

	c := make(chan syncTaskStatus)
	go func() {
		a.syncActive.Store(true)
		res, err := a.SyncKeychains(ctx)
		if err != nil {
			c <- syncTaskStatus{
				Finished: true,
				Error:    fmt.Errorf("sync error: %w", err),
			}
		} else {
			c <- syncTaskStatus{
				Finished:       true,
				SyncTaskResult: *res,
			}
		}
		close(c)
		a.syncActive.Store(false)
	}()

	return c, nil
}
