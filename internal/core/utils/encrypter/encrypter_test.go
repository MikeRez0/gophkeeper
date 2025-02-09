package encrypter_test

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/core/utils/encrypter"
	"github.com/stretchr/testify/assert"
)

func setup() error {
	return nil
}

func shutdown() {
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

	enc, err := encrypter.NewEncrypter(l.Named("encrypter"), 32)
	assert.NoError(t, err)
	dec, err := encrypter.NewDecrypter(l.Named("decrypter"), 32)
	assert.NoError(t, err)

	testPass := "mypassword"

	t.Run("base", func(t *testing.T) {
		testData := []byte("MY SECRET DATA")

		env, err := enc.Encrypt(testData, []byte(testPass))
		assert.NoError(t, err)

		data, err := dec.Decrypt(env, []byte(testPass))
		assert.NoError(t, err)
		assert.NotEqual(t, testData, env.Data)

		assert.Equal(t, testData, data)
	})

	t.Run("fail key", func(t *testing.T) {
		testData := "MY SECRET DATA"

		env, err := enc.Encrypt([]byte(testData), []byte(testPass))
		assert.NoError(t, err)

		env.Key = []byte("BADKEY")

		_, err = dec.Decrypt(env, []byte(testPass))
		assert.Error(t, err)
	})

	t.Run("big data", func(t *testing.T) {
		testData := make([]byte, 1024*1024*1024)
		bigPass := make([]byte, 1024)
		_, err := rand.Read(testData)
		assert.NoError(t, err)
		_, err = rand.Read(bigPass)
		assert.NoError(t, err)

		env, err := enc.Encrypt(testData, bigPass)
		assert.NoError(t, err)

		data, err := dec.Decrypt(env, bigPass)
		assert.NoError(t, err)
		assert.NotEqual(t, testData, env.Data)

		assert.Equal(t, testData, data)
	})
}
