package keychain

import (
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/utils/encrypter"
	"go.uber.org/zap"
)

type Keychain struct {
	data    *domain.KCData
	Pass    string
	Items   []*KeychainItem
	IsDirty bool
	keySize uint
	enc     *encrypter.Encrypter
	dec     *encrypter.Decrypter
	log     *zap.Logger
}

const cKeySize = 32

func NewKeychain(data *domain.KCData, log *zap.Logger) (*Keychain, error) {
	if data == nil {
		data = &domain.KCData{}
	}
	enc, err := encrypter.NewEncrypter(log, cKeySize)
	if err != nil {
		return nil, fmt.Errorf("error creating encrypter: %w", err)
	}
	dec, err := encrypter.NewDecrypter(log, cKeySize)
	if err != nil {
		return nil, fmt.Errorf("error creating decrypter: %w", err)
	}

	return &Keychain{
		data:    data,
		keySize: cKeySize,
		log:     log,
		enc:     enc,
		dec:     dec,
	}, nil
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

	env, err := kc.enc.Encrypt(secret, []byte(kc.Pass))
	if err != nil {
		return fmt.Errorf("error storing secret:%w", err)
	}

	item.data.Value = env.Data
	item.data.Key = env.Key

	return nil
}

func (kc *Keychain) GetSecret(item *KeychainItem) ([]byte, error) {

	secret, err := kc.dec.Decrypt(&encrypter.Envelope{
		Key:  item.data.Key,
		Data: item.data.Value,
	}, []byte(kc.Pass))

	if err != nil {
		return nil, fmt.Errorf("decryption failed:%w", err)
	}

	return secret, nil
}

func (kc *Keychain) Data() *domain.KCData {
	return kc.data
}
