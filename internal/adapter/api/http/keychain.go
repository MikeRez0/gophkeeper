package http

import (
	"fmt"
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

type KeychainHandler struct {
	Handler
	service port.IKeychainDataService
}

func NewKeychainHandler(service port.IKeychainDataService, logger *zap.Logger) (*KeychainHandler, error) {
	return &KeychainHandler{
		Handler: Handler{logger: logger},
		service: service}, nil
}

func (h *KeychainHandler) GetKeychainList(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	list, err := h.service.KeychainList(ctx, payload.UserID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, list)
}

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

	k, err := h.service.KeychainCreate(ctx, &domain.KCData{
		ID:      keychainID,
		OwnerID: payload.UserID,
	})
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, k)
}

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

type keychainItemStruct struct {
	Label    *string                  `json:"label"`
	MetaData *domain.KeychainItemMeta `json:"meta"`
	Value    *[]byte                  `json:"value"`
	Key      *[]byte                  `json:"key"`
	ItemType *domain.KCItemType       `json:"type"`
	Created  time.Time
	Changed  time.Time
}

func (h *KeychainHandler) SaveKeychainItem(ctx *gin.Context) {
	payload := getAuthPayload(ctx)

	var item domain.KCItemData

	if u, err := parseUUIDParam(ctx, cKeychainParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		item.KeyChainID = domain.KeychainID(u)
	}
	if u, err := parseUUIDParam(ctx, cKeychainItemParamName); err != nil {
		h.handleValidationError(ctx, err)
		return
	} else {
		item.ID = domain.KeychainItemID(u)
	}

	var reqItem keychainItemStruct

	err := ctx.ShouldBindBodyWithJSON(&reqItem)
	if err != nil {
		h.handleValidationError(ctx, err)
		return
	}

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

	pItem, err := h.service.KeychainSaveItem(ctx, payload.UserID, &item)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	h.handleSuccess(ctx, pItem)
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
