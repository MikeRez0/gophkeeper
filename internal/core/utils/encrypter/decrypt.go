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

type Decrypter struct {
	log     *zap.Logger
	keySize uint
}

func NewDecrypter(log *zap.Logger, keySize uint) (*Decrypter, error) {
	return &Decrypter{
		keySize: keySize,
		log:     log,
	}, nil
}

func (d *Decrypter) Decrypt(envelope *Envelope, pass []byte) ([]byte, error) {
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
