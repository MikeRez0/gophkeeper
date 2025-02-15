package keychain

import (
	"bytes"
	"fmt"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/utils/encrypter"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Keychain struct {
	Pass     string
	enc      *encrypter.Encrypter
	dec      *encrypter.Decrypter
	log      *zap.Logger
	data     *domain.KCData
	Items    []*KeychainItem
	changed  bool
	keySize  uint
	SyncTime time.Time
}

const cKeySize = 32

func NewKeychain(data *domain.KCData, log *zap.Logger) (*Keychain, error) {
	newKeychain := false
	if data == nil {
		data = &domain.KCData{
			Name: "My keychain",
			ID:   domain.KeychainID(uuid.New()),
		}
		newKeychain = true
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
		changed: newKeychain,
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

func (kc *Keychain) ApplyItemFromData(data *domain.KCItemData) *KeychainItem {
	for _, item := range kc.Items {
		if item.data.ID == data.ID {
			item.data = data
			item.changed = false
			return item
		}
	}
	item := newKeychainItemFromData(data)
	kc.Items = append(kc.Items, item)
	return item
}

func (kc *Keychain) StoreSecret(item *KeychainItem, secret []byte) error {
	if oldSecret, err := kc.GetSecret(item); err != nil {
		return err
	} else if bytes.Equal(oldSecret, secret) {
		return nil
	}

	env, err := kc.enc.Encrypt(secret, []byte(kc.Pass))
	if err != nil {
		return fmt.Errorf("error storing secret:%w", err)
	}

	item.data.Value = env.Data
	item.data.Key = env.Key

	item.touch()

	return nil
}

func (kc *Keychain) GetSecret(item *KeychainItem) ([]byte, error) {
	if len(item.data.Value) == 0 {
		return item.data.Value, nil
	}
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

func (kc *Keychain) IsChanged() bool {
	return kc.changed
}
