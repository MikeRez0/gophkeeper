package keychain

import (
	"fmt"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/google/uuid"
)

type KeychainItem struct {
	data       *domain.KCItemData
	lastChange time.Time
	changed    bool
}

func newKeychainItem(kc *Keychain, itemType domain.KCItemType) *KeychainItem {
	item := KeychainItem{
		data: newKeychainItemData(kc.data.ID, itemType),
	}

	item.touch()

	return &item
}
func newKeychainItemFromData(data *domain.KCItemData) *KeychainItem {
	item := KeychainItem{
		data: data,
	}

	return &item
}

func (ki *KeychainItem) MetaData() domain.KeychainItemMeta {
	return ki.data.MetaData
}

func (ki *KeychainItem) MetaDataItem(key string) string {
	if v, ok := ki.data.MetaData[key]; ok {
		return v
	}
	return ""
}

func (ki *KeychainItem) MetaDataSetItem(key string, value string) {
	ki.data.MetaData[key] = value

	ki.touch()
}

func (ki *KeychainItem) touch() {
	ki.changed = true
	ki.lastChange = time.Now()
	ki.data.ClientTime = ki.lastChange
}

func newKeychainItemData(keychainID domain.KeychainID, itemType domain.KCItemType) *domain.KCItemData {
	metas := make(domain.KeychainItemMeta, 5)
	metas[domain.KCMetaKeyComment] = ""
	switch itemType { //nolint:exhaustive // not all types have default meta values
	case domain.KCItemTypePassword:
		metas[domain.KCMetaKeyLogin] = ""
		metas[domain.KCMetaKeySite] = ""
	case domain.KCItemTypeCardNumber:
		metas[domain.KCMetaKeyIssuer] = ""
		metas[domain.KCMetaKeyOwner] = ""
		metas[domain.KCMetaKeyValidTo] = ""
	}

	return &domain.KCItemData{
		KeyChainID: keychainID,
		ID:         domain.KeychainItemID(uuid.New()),
		ItemType:   itemType,
		MetaData:   metas,
	}
}

func (ki *KeychainItem) String() string {
	return fmt.Sprintf(`Item: %s
Metadata: %v`,
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
