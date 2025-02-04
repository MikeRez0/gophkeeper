package keychain

import (
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

type Keychain struct {
	data    *domain.KCData
	Pass    string
	Items   []*KeychainItem
	IsDirty bool
	keySize uint
}

const cKeySize = 256

func NewKeychain(data *domain.KCData) *Keychain {
	if data == nil {
		data = &domain.KCData{}
	}
	return &Keychain{
		data:    data,
		keySize: cKeySize,
	}
}

func (kc *Keychain) KeySize() uint {
	return kc.keySize
}

func (kc *Keychain) NewItem(itemType domain.KCItemType) *KeychainItem {
	item := newKeychainItem(kc, itemType)
	kc.Items = append(kc.Items, item)
	return item
}

func (kc *Keychain) AppendItemFromData(data *domain.KCItemData) *KeychainItem {
	item := newKeychainItemFromData(data)
	kc.Items = append(kc.Items, item)
	return item
}

func (kc *Keychain) StoreSecret(item *KeychainItem, secret []byte) error {
	//TODO: Encryption
	item.data.Value = secret
	item.data.Key = []byte(kc.Pass)

	return nil
}

func (kc *Keychain) GetSecret(item *KeychainItem) ([]byte, error) {
	var secret []byte
	//TODO: Decryption
	if kc.Pass == string(item.data.Key) {
		secret = item.data.Value
	} else {
		return nil, fmt.Errorf("Decryption failed")
	}

	return secret, nil
}
