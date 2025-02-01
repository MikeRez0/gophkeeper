package signer_test

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/core/utils/signer"
	"github.com/stretchr/testify/assert"
)

const cPubKey string = `-----BEGIN RSA PUBLIC KEY-----
MIIECgKCBAEAronU2RrZ88VNCedYnHnJSIP8NB5jJS23n3ukoZVW3ZXG0E1pa7qQ
Shj1iP6AaO7tH1eEjkv/HWZ6d/qF/setN9hM8rvf/+daDkHaexqKa7TDuMYCcxqa
XEzXXE7UeFJzJbHwphm5/I7TyNIDK298P50sDm9R6JD1DDqoydgiANHFa5aflNfq
IyAc3d6P3N024BidS3ODz++RJORI1WkCBupF5uLmPb1KkJmpugS9tFOFs4q1GyDT
y9DcIOW9Z5Yax3mky0lvfNEp47NHnqu/Qgn5fm3ntpc2ahxZIcbWGG5qWUJKfh3x
NsOEKl8/sxTNJHBNFEsC2qbhH0P2HHs7yQRjLNyd31AFRMbvxLNCS3w2+JMV/YIP
M5i8+6kkjNf6ST+rZpVN4MBnvfwrx7hr3sh3QJ6UWHJ24vWvmtZI+evjz7rF3+gE
4Joe3s2+0GAsU1R9fvuUwfBgc7C0CxrB5oXVfISCvyrnt9OuyPZ+uPG0uRXfg1WL
+p2jPUPalcUQy75Z1hMbILbWwCvNifNtFMDjxbX46GTB2C7yRkaCYVcSjtfWm74r
Y4vFm2APy/TbVfMk/9oG1TrIf9tRpg80KOkc70C1FSJh3eRhfEZX/SZpaAR+XRva
8z5bKVOF0mixVm1tkWjdzyoL5OGJV0I7G32L5X9JtMCTTBTBhrpZqJoXo2FEOp/n
CFueYOCV5wCZeDqL5+WekbN52NQkInH9kxLOo+kWtiYuc2Pbqn9R5Prs3GpxMXvu
r4oq2s1qWDAI2WMPC/R7SePC6j9dbLejOHrgHthy/m9DHOK3GW8UQ5vheVrp8LjR
kiBOHDZ1pbDYgPym8u3Dlp6dVfp9pLpqi9OoXTI8+5puYxah5IyW1XUSRK/rRP1t
QlsKd86ef493fU0h5pJKKliw3rIxYzDXtvUO9tWRJWLyb49xNhoAIQ0l3W+pf1ix
DEXRcTbrhPRJMkw8czcqbjBrlaxA15Fod8C/jZZQiLNweCBd/lWXcQwZB+1pnwvk
cXs2N5g8NKOvpTctu/LGqHMFYxfiTuHUrZTXJbtriiEZm06RefSndgjFbHO3MOGP
TXCGvWyVyufAFE/988WtXBVn9ycnAbE2uNDxprDgfzY0JY2knA5KkyRes7EOPzAh
wqj9wfdAzwlUnM7T5hVyfhSff4vT1aZ2n1oQwljv9XLVbt2UxM7jQXg+KpuoGM7e
rnkUDB/WXsBOv1ZIH4bwdXT/ovadEA7OmRAQMi8z2BvfxtoqRvtrC0xab8mVKqz2
TRxEvuujj1V51tc8axJGl97GJTz0V5vQf8VDkOUyCBfZaWXIRZspCbSbIrzCVCla
BixQ7CSgPyKqiyH10UMZacUyiwc3YzvyOwIDAQAB
-----END RSA PUBLIC KEY-----
`

const cPrivateKey string = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

const (
	cPubKeyFilename     = "pubkey.pem"
	cPrivateKeyFilename = "key.pem"
)

func setup() error {
	fpub, err := os.Create(cPubKeyFilename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer fpub.Close() //nolint:errcheck // make it simpler

	_, err = fpub.WriteString(cPubKey)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	fpriv, err := os.Create(cPrivateKeyFilename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer fpriv.Close() //nolint:errcheck // make it simpler

	_, err = fpriv.WriteString(cPrivateKey)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func shutdown() {
	err := os.Remove(cPubKeyFilename)
	if err != nil {
		log.Println(err)
	}
	err = os.Remove(cPrivateKeyFilename)
	if err != nil {
		log.Println(err)
	}
}

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		shutdown()
		os.Exit(1)
	}
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestEncrypt(t *testing.T) {
	l := logger.NewLogger(&config.App{LogLevel: "debug"})

	enc, err := signer.NewEncrypter(cPubKeyFilename, l.Named("encrypter"))
	assert.NoError(t, err)
	dec, err := signer.NewDecrypter(cPrivateKeyFilename, l.Named("decrypter"))
	assert.NoError(t, err)

	t.Run("base", func(t *testing.T) {
		testData := []byte("MY SECRET DATA")

		env, err := enc.Encrypt(testData)
		assert.NoError(t, err)

		data, err := dec.Decrypt(env)
		assert.NoError(t, err)
		assert.NotEqual(t, testData, env.Data)

		assert.Equal(t, testData, data)
	})

	t.Run("fail key", func(t *testing.T) {
		testData := "MY SECRET DATA"

		env, err := enc.Encrypt([]byte(testData))
		assert.NoError(t, err)

		env.Key = []byte("BADKEY")

		_, err = dec.Decrypt(env)
		assert.Error(t, err)
	})

	t.Run("big data", func(t *testing.T) {
		testData := make([]byte, 1024*1024*1024)
		_, err := rand.Read(testData)
		assert.NoError(t, err)

		env, err := enc.Encrypt(testData)
		assert.NoError(t, err)

		data, err := dec.Decrypt(env)
		assert.NoError(t, err)
		assert.NotEqual(t, testData, env.Data)

		assert.Equal(t, testData, data)
	})
}

func TestSigner(t *testing.T) {
	s := signer.NewSigner("MYKEY")

	t.Run("json hash", func(t *testing.T) {
		data := struct {
			name string
			num  int
		}{name: "test", num: 1}

		val, err := s.GetHashJSON(data)
		assert.NoError(t, err)
		assert.True(t, s.ValidateJSON(data, val))

		val += "BAD"
		assert.False(t, s.ValidateJSON(data, val))
	})

	t.Run("byte array hash", func(t *testing.T) {
		data := []byte("MY DATA")

		val, err := s.GetHashBA(data)
		assert.NoError(t, err)
		assert.True(t, s.Validate(data, val))

		val += "BAD"
		assert.False(t, s.Validate(data, val))
	})
}
