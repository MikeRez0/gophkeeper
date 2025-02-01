package signer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
)

type Decrypter struct {
	key *rsa.PrivateKey
	log *zap.Logger
}

func NewDecrypter(keyFilename string, log *zap.Logger) (*Decrypter, error) {
	block, err := readPem(keyFilename)
	if err != nil {
		return nil, fmt.Errorf("error reading key: %w", err)
	}
	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing private key: %w", err)
		}
		return &Decrypter{key: privateKey, log: log}, nil
	} else {
		return nil, errors.New("rsa private key is requiered")
	}
}

func (d *Decrypter) Decrypt(envelope *Envelope) ([]byte, error) {
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, d.key, envelope.Key, []byte{0})
	if err != nil {
		return nil, fmt.Errorf("key decrypting error: %w", err)
	}

	gcmCipher, err := createCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	nonce := envelope.Data[:gcmCipher.NonceSize()]

	data, err := gcmCipher.Open(nil, nonce, envelope.Data[gcmCipher.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting: %w", err)
	}

	return data, nil
}

func readPem(filename string) (*pem.Block, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening crypto-key file: %w", err)
	}
	defer func() { _ = f.Close() }()

	key, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading crypto-key file: %w", err)
	}

	block, _ := pem.Decode(key)
	return block, nil
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
