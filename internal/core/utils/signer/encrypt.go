package signer

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type Envelope struct {
	Key  []byte
	Data []byte
}

type Encrypter struct {
	key *rsa.PublicKey
	log *zap.Logger
}

func NewEncrypter(keyFilename string, log *zap.Logger) (*Encrypter, error) {
	block, err := readPem(keyFilename)
	if err != nil {
		return nil, fmt.Errorf("error reading key: %w", err)
	}
	if block.Type == "RSA PUBLIC KEY" {
		publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing public key: %w", err)
		}
		return &Encrypter{key: publicKey, log: log}, nil
	} else {
		return nil, errors.New("rsa public key is requiered")
	}
}

func (e *Encrypter) Encrypt(data []byte) (*Envelope, error) {
	aesKey, err := getRandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("error create aes key: %w", err)
	}

	encrData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.key, aesKey, []byte{0})
	if err != nil {
		return nil, fmt.Errorf("error encrypting request: %w", err)
	}

	gcmCipher, err := createCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	nonce, err := getRandomBytes(gcmCipher.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	myData := gcmCipher.Seal(nonce, nonce, data, nil)

	env := &Envelope{
		Data: myData,
		Key:  encrData,
	}

	return env, nil
}
