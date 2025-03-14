// Package keychain implements Keychain item business logic (ex.: encrypt, decrypt).
package keychain

import (
	"fmt"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/google/uuid"
)

// KeychainItem stores item data and changed status.
type KeychainItem struct {
	data       *domain.KCItemData
	lastChange time.Time
	changed    bool
}

// NewKeychainItem create new keyhcain item with [itemType].
func NewKeychainItem(kc domain.KeychainID, itemType domain.KCItemType) *KeychainItem {
	item := KeychainItem{
		data: newKeychainItemData(kc, itemType),
	}

	item.touch()

	return &item
}

// NewKeychainItemFromData create new keyhcain item from provided data.
func NewKeychainItemFromData(data *domain.KCItemData) *KeychainItem {
	item := KeychainItem{
		data: data,
	}

	return &item
}

// MetaData returns meta data values.
func (ki *KeychainItem) MetaData() domain.KeychainItemMeta {
	return ki.data.MetaData
}

// MetaDataItem returns meta data value with given name.
func (ki *KeychainItem) MetaDataItem(key string) string {
	if v, ok := ki.data.MetaData[key]; ok {
		return v
	}
	return ""
}

// MetaDataSetItem sets meta data value with given name.
func (ki *KeychainItem) MetaDataSetItem(key string, value string) {
	ki.data.MetaData[key] = value

	ki.touch()
}

func (ki *KeychainItem) touch() {
	ki.changed = true
	ki.lastChange = time.Now().UTC()
	ki.data.ClientTime = ki.lastChange
}

// fillMetaData controls meta data with item type.
func fillMetaData(item *domain.KCItemData) {
	if item.MetaData == nil {
		item.MetaData = make(domain.KeychainItemMeta, 5)
	}

	for k, v := range item.MetaData {
		if v == "" {
			delete(item.MetaData, k)
		}
	}

	if _, ok := item.MetaData[domain.KCMetaKeyComment]; !ok {
		item.MetaData[domain.KCMetaKeyComment] = ""
	}
	switch item.ItemType { //nolint:exhaustive // not all cases needed
	case domain.KCItemTypePassword:
		if _, ok := item.MetaData[domain.KCMetaKeyLogin]; !ok {
			item.MetaData[domain.KCMetaKeyLogin] = ""
		}
		if _, ok := item.MetaData[domain.KCMetaKeySite]; !ok {
			item.MetaData[domain.KCMetaKeySite] = ""
		}
	case domain.KCItemTypeCardNumber:
		if _, ok := item.MetaData[domain.KCMetaKeyIssuer]; !ok {
			item.MetaData[domain.KCMetaKeyIssuer] = ""
		}
		if _, ok := item.MetaData[domain.KCMetaKeyOwner]; !ok {
			item.MetaData[domain.KCMetaKeyOwner] = ""
		}
		if _, ok := item.MetaData[domain.KCMetaKeyValidTo]; !ok {
			item.MetaData[domain.KCMetaKeyValidTo] = ""
		}
	}
}

func newKeychainItemData(keychainID domain.KeychainID, itemType domain.KCItemType) *domain.KCItemData {
	item := domain.KCItemData{
		KeyChainID: keychainID,
		ID:         domain.KeychainItemID(uuid.New()),
		ItemType:   itemType,
	}

	fillMetaData(&item)

	return &item
}

func (ki *KeychainItem) String() string {
	return fmt.Sprintf("Item: %s Metadata: [%v]\n",
		ki.data.Label, ki.MetaData())
}

func (ki *KeychainItem) Label() string {
	return ki.data.Label
}

func (ki *KeychainItem) SetLabel(label string) {
	ki.data.Label = label
	ki.touch()
}

func (ki *KeychainItem) Data() *domain.KCItemData {
	return ki.data
}

func (ki *KeychainItem) IsChanged() bool {
	return ki.changed
}

// SetType changes type of item.
func (ki *KeychainItem) SetType(t domain.KCItemType) {
	ki.data.ItemType = t
	fillMetaData(ki.data)
	ki.touch()
}

// StoreSecret stores secret value and key.
func (ki *KeychainItem) StoreSecret(key, value []byte) {
	ki.data.Key = key
	ki.data.Value = value
	ki.touch()
}
