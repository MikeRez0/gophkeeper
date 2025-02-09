package encrypter

import (
	"fmt"

	"go.uber.org/zap"
)

type Envelope struct {
	Key  []byte
	Data []byte
}

type Encrypter struct {
	log     *zap.Logger
	keySize uint
}

func NewEncrypter(log *zap.Logger, keySize uint) (*Encrypter, error) {
	return &Encrypter{
		keySize: keySize,
		log:     log,
	}, nil
}

func (e *Encrypter) Encrypt(data []byte, pass []byte) (*Envelope, error) {
	passKey := getPassKey(pass)

	aesKey, err := getRandomBytes(int(e.keySize))
	if err != nil {
		return nil, fmt.Errorf("error create aes key: %w", err)
	}

	keyCipher, err := createCipher(passKey)
	if err != nil {
		return nil, fmt.Errorf("error creating key cipher: %w", err)
	}
	nonce, err := getRandomBytes(keyCipher.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}
	encrKey := keyCipher.Seal(nonce, nonce, aesKey, nil)

	dataCipher, err := createCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating data cipher: %w", err)
	}

	nonce, err = getRandomBytes(dataCipher.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	encrData := dataCipher.Seal(nonce, nonce, data, nil)

	env := &Envelope{
		Data: encrData,
		Key:  encrKey,
	}

	return env, nil
}
