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
	ki.lastChange = time.Now().UTC()
	ki.data.ClientTime = ki.lastChange
}

func fillMetaData(item *domain.KCItemData) {
	if item.MetaData == nil {
		item.MetaData = make(domain.KeychainItemMeta, 5)
	}
	if _, ok := item.MetaData[domain.KCMetaKeyComment]; !ok {
		item.MetaData[domain.KCMetaKeyComment] = ""
	}
	switch item.ItemType {
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

func (ki *KeychainItem) SetType(t domain.KCItemType) {
	ki.data.ItemType = t
	fillMetaData(ki.data)
	ki.touch()
}
