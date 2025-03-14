// Package encrypter contains cryptography utils.
package encrypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// Envelope structure for encrypting data with key.
type Envelope struct {
	Key  []byte
	Data []byte
}

// Crypter is object for encrypting/decrypting secret data to/from envelope.
type Crypter struct {
	log     *zap.Logger
	keySize uint
}

// NewCrypter creates new Crypter.
func NewCrypter(log *zap.Logger, keySize uint) (*Crypter, error) {
	return &Crypter{
		keySize: keySize,
		log:     log,
	}, nil
}

// Encrypt returns Envelope with encrypted data by password.
func (c *Crypter) Encrypt(data []byte, pass []byte) (*Envelope, error) {
	passKey := getPassKey(pass)

	aesKey, err := getRandomBytes(int(c.keySize))
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

// Decrypt trys to decrypt envelope with password, return secret data if all ok.
func (c *Crypter) Decrypt(envelope *Envelope, pass []byte) ([]byte, error) {
	passKey := getPassKey(pass)

	keyCipher, err := createCipher(passKey)
	if err != nil {
		return nil, fmt.Errorf("error creating key cipher: %w", err)
	}
	if len(envelope.Key) < keyCipher.NonceSize() {
		return nil, fmt.Errorf("error parsing key: %w", errors.New("bad key size"))
	}
	nonce := envelope.Key[:keyCipher.NonceSize()]
	aesKey, err := keyCipher.Open(nil, nonce, envelope.Key[keyCipher.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting key: %w", err)
	}

	dataCipher, err := createCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	nonce = envelope.Data[:dataCipher.NonceSize()]

	data, err := dataCipher.Open(nil, nonce, envelope.Data[dataCipher.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data: %w", err)
	}

	return data, nil
}

func createCipher(key []byte) (cipher.AEAD, error) {
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating aes cipher: %w", err)
	}
	gcmCipher, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, fmt.Errorf("error creating gcm cipher: %w", err)
	}
	return gcmCipher, nil
}

func getRandomBytes(n int) ([]byte, error) {
	data := make([]byte, n)

	_, err := rand.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error generate random: %w", err)
	}

	return data, nil
}

func getPassKey(pass []byte) []byte {
	hashp := sha256.Sum256(pass)
	passKey := make([]byte, 32)
	copy(passKey, hashp[:])

	return passKey
}
