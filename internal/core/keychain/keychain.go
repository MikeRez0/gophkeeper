package keychain

import (
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

type KeyChain struct {
	Data    *domain.KCData
	Pass    string
	Items   []*KeyChainItem
	IsDirty bool
	KeySize uint
}

type KeyChainItem struct {
	Data     *domain.KCItemData
	KeyChain *KeyChain
	Changed  bool
}

const KeySize = 256

func NewKeyChain(data *domain.KCData) *KeyChain {
	if data == nil {
		data = &domain.KCData{}
	}
	return &KeyChain{
		Data:    data,
		KeySize: KeySize,
	}
}

func (kc *KeyChain) NewItem(itemType domain.KCItemType) *KeyChainItem {
	item := KeyChainItem{
		Data:     newKeyChainItemData(kc.Data.ID, itemType),
		KeyChain: kc,
		Changed:  true,
	}
	kc.Items = append(kc.Items, &item)

	return &item
}

func (kc *KeyChain) AppendItemFromData(data *domain.KCItemData) *KeyChainItem {
	item := KeyChainItem{
		Data:     data,
		KeyChain: kc,
		Changed:  false,
	}
	kc.Items = append(kc.Items, &item)

	return &item
}

func (kc *KeyChain) StoreSecret(item *KeyChainItem, secret []byte) error {
	//TODO: Encryption
	item.Data.Value = secret
	item.Data.Key = []byte(kc.Pass)

	return nil
}

func (kc *KeyChain) GetSecret(item *KeyChainItem) ([]byte, error) {
	var secret []byte
	//TODO: Decryption
	if kc.Pass == string(item.Data.Key) {
		secret = item.Data.Value
	} else {
		return nil, fmt.Errorf("Decryption failed")
	}

	return secret, nil
}

func newKeyChainItemData(keychainID domain.KeyChainID, itemType domain.KCItemType) *domain.KCItemData {
	metas := make(domain.KeyChainItemMeta, 5)
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
		ItemType:   itemType,
		MetaData:   metas,
	}
}
