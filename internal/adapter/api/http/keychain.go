package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	cKeychainParamName     = "keychain_id"
	cKeychainItemParamName = "item_id"
)

// KeychainHandler - handler for keychain requests.
type KeychainHandler struct {
	Handler
	service port.IKeychainDataService
}

// NewKeychainHandler - creates new keychain handler.
func NewKeychainHandler(service port.IKeychainDataService, logger *zap.Logger) (*KeychainHandler, error) {
	return &KeychainHandler{
		Handler: Handler{logger: logger},
		service: service}, nil
}

// GetKeychainList - get keychain list for authenticated user.
func (h *KeychainHandler) GetKeychainList(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	list, err := h.service.KeychainList(ctx, payload.UserID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, list)
}

type keychainStruct struct {
	Name string `json:"name"`
}

// SaveKeychain - save keychain for authenticated user (updates if keychainID is provided).
func (h *KeychainHandler) SaveKeychain(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	var (
		keychainID domain.KeychainID
	)

	if u, err := parseUUIDParam(ctx, cKeychainParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		keychainID = domain.KeychainID(u)
	}

	keychainReq := &keychainStruct{}
	if err := ctx.ShouldBindBodyWithJSON(keychainReq); err != nil {
		h.handleValidationError(ctx, err)
		return
	}

	k, err := h.service.KeychainSave(ctx, payload.UserID, &domain.KCData{
		ID:      keychainID,
		OwnerID: payload.UserID,
		Name:    keychainReq.Name,
	})
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, k)
}

// GetKeychain - read keychain for authenticated user.
func (h *KeychainHandler) GetKeychain(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	var (
		keychainID domain.KeychainID
	)

	if u, err := parseUUIDParam(ctx, cKeychainParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		keychainID = domain.KeychainID(u)
	}

	k, err := h.service.KeychainGet(ctx, payload.UserID, keychainID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, k)
}

// ListKeychainItems - list all keychain items for authenticated user.
func (h *KeychainHandler) ListKeychainItems(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	var (
		keychainID domain.KeychainID
	)

	if u, err := parseUUIDParam(ctx, cKeychainParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		keychainID = domain.KeychainID(u)
	}

	items, err := h.service.KeychainGetItemsSince(ctx, payload.UserID, keychainID, time.Time{})
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, items)
}

// keychainItemStruct structure for json-requested item.
type keychainItemStruct struct {
	Label      *string                  `json:"label"`
	MetaData   *domain.KeychainItemMeta `json:"meta"`
	Value      *[]byte                  `json:"value"`
	Key        *[]byte                  `json:"key"`
	ItemType   *domain.KCItemType       `json:"type"`
	ID         *domain.KeychainItemID   `json:"id"`
	ClientTime *time.Time               `json:"client_time"`
}

// SaveKeychainItem - save (create/update) keychain item for authenticated user.
func (h *KeychainHandler) SaveKeychainItem(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	item := &domain.KCItemData{}

	if u, err := parseUUIDParam(ctx, cKeychainParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		item.KeyChainID = domain.KeychainID(u)
	}

	var reqItem keychainItemStruct

	err := ctx.ShouldBindBodyWithJSON(&reqItem)
	if err != nil {
		h.handleValidationError(ctx, err)
		return
	}

	parseItemData(&reqItem, item)

	if u, err := parseUUIDParam(ctx, cKeychainItemParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		item.ID = domain.KeychainItemID(u)
	}

	item, updated, err := h.service.KeychainSaveItem(ctx, payload.UserID, item)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	if updated {
		h.handleSuccess(ctx, item)
	} else {
		h.handleSuccessWithStatus(ctx, item, http.StatusFound)
	}
}

func parseUUIDParam(ctx *gin.Context, name string) (uuid.UUID, error) {
	if str, ok := ctx.Params.Get(name); ok {
		if i, err := uuid.Parse(str); err == nil {
			return i, nil
		} else {
			return uuid.Nil, fmt.Errorf("error parsing uuid: %w", err)
		}
	} else {
		return uuid.Nil, fmt.Errorf("%s is required", name)
	}
}

func parseItemData(reqItem *keychainItemStruct, item *domain.KCItemData) {
	if reqItem.Label != nil {
		item.Label = *reqItem.Label
	}
	if reqItem.Value != nil {
		item.Value = *reqItem.Value
	}
	if reqItem.Key != nil {
		item.Key = *reqItem.Key
	}
	if reqItem.ItemType != nil {
		item.ItemType = *reqItem.ItemType
	}
	if reqItem.MetaData != nil {
		item.MetaData = *reqItem.MetaData
	}
	if !reqItem.ClientTime.IsZero() {
		item.ClientTime = *reqItem.ClientTime
	}
	if reqItem.ID != nil {
		item.ID = *reqItem.ID
	}
}

// Sync - runs synchronisation process of keychain for authenticated user.
func (h *KeychainHandler) Sync(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	var (
		keychainID domain.KeychainID
	)

	if u, err := parseUUIDParam(ctx, cKeychainParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		keychainID = domain.KeychainID(u)
	}

	var fromTime time.Time

	stime := ctx.Query("from_time")
	if stime != "" {
		if t, err := time.Parse(time.RFC3339, stime); err == nil {
			fromTime = t
		} else {
			h.handleValidationError(ctx, fmt.Errorf("error parsing time: %w", err))
			return
		}
	}

	var itemsReq []keychainItemStruct

	err := ctx.ShouldBindBodyWithJSON(&itemsReq)
	if err != nil {
		h.handleValidationError(ctx, err)
		return
	}

	items := make([]*domain.KCItemData, 0, len(itemsReq))
	serverTime := time.Now()
	for _, itemReq := range itemsReq {
		item := domain.KCItemData{}
		item.KeyChainID = keychainID
		parseItemData(&itemReq, &item)
		item.ServerTime = serverTime
		items = append(items, &item)
	}

	result, err := h.service.Sync(ctx, payload.UserID, keychainID, fromTime, items)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, result)
}

// GetKeychainItem - reads keychain item for authenticated user.
func (h *KeychainHandler) GetKeychainItem(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	var (
		keychainID domain.KeychainID
		itemID     domain.KeychainItemID
	)

	if u, err := parseUUIDParam(ctx, cKeychainParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		keychainID = domain.KeychainID(u)
	}

	if u, err := parseUUIDParam(ctx, cKeychainItemParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		itemID = domain.KeychainItemID(u)
	}

	item, err := h.service.KeychainGetItem(ctx, payload.UserID, keychainID, itemID)
	if err != nil {
		h.handleAbort(ctx, err)
	}

	h.handleSuccess(ctx, item)
}
