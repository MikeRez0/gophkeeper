// Package app contains client GophKeeper application.
package app

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/keychain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/MikeRez0/gophkeeper/internal/core/utils/encrypter"
	"go.uber.org/zap"
)

// ClientApp is the core object of client application.
// It's consist of:
//   - configuration parameters
//   - cryptography objects (encrypter, decrypter)
//   - local storage service
//
// Most usecases of app must instantiate the ClientApp, warm (bootstrap) it and then execute business logic.
type ClientApp struct {
	config       *config.ConfigClient
	enc          *encrypter.Crypter
	Log          *zap.Logger
	httpClient   *http.Client
	Service      port.IKeychainDataService
	syncTime     time.Time
	token        string
	SyncInterval time.Duration
	UserID       domain.UserID
	syncActive   atomic.Bool
}

const cKeySize = 32

// NewApp creates new client app object.
func NewApp(conf *config.ConfigClient, service port.IKeychainDataService, log *zap.Logger) (*ClientApp, error) {
	caCertPool := x509.NewCertPool()
	if conf.TLSCertFile != "" {
		caCert, err := os.ReadFile(conf.TLSCertFile)
		if err != nil {
			return nil, fmt.Errorf("read cert file error: %w", err)
		}

		caCertPool.AppendCertsFromPEM(caCert)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	enc, err := encrypter.NewCrypter(log, cKeySize)
	if err != nil {
		return nil, fmt.Errorf("error creating encrypter: %w", err)
	}

	return &ClientApp{
		config:       conf,
		SyncInterval: conf.SyncInterval,
		Log:          log,
		httpClient:   client,
		enc:          enc,
		Service:      service,
		syncTime:     time.Time{},
		UserID:       domain.UserID(0),
	}, nil
}

// SetToken saves token for app.
func (a *ClientApp) SetToken(token string) {
	a.token = token
}

// Connect trys to connect to server, saves token if success.
func (a *ClientApp) Connect(ctx context.Context, login, password string) error {
	data, err := a.doRequest(ctx,
		"/api/user/login",
		http.MethodPost,
		map[string]string{"login": login, "password": password})
	if err != nil {
		return fmt.Errorf("error on login: %w", err)
	}

	payload := struct {
		Token string `json:"token"`
	}{}
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("error parsing response %w", err)
	}

	a.token = payload.Token

	return nil
}

// RegisterUser trys to register new user on server, saves token if success.
func (a *ClientApp) RegisterUser(ctx context.Context, login, password string) error {
	data, err := a.doRequest(ctx,
		"/api/user/register",
		http.MethodPost,
		map[string]string{"login": login, "password": password})
	if err != nil {
		return fmt.Errorf("error on register: %w", err)
	}

	payload := struct {
		Token string `json:"token"`
	}{}
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("error parsing response %w", err)
	}

	a.token = payload.Token

	return nil
}

// doRequest - common method for make request to server.
func (a *ClientApp) doRequest(ctx context.Context, path string, method string, data any) ([]byte, error) {
	requestStr := a.config.HostString + path

	var (
		body []byte
		err  error
	)
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("error creating req body %w", err)
		}
	}

	req, err := http.NewRequest(method, requestStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error on %s : %w", requestStr, err)
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	if a.token != "" {
		req.Header.Add("Authorization", "Bearer "+a.token)
	}

	req = req.WithContext(ctx)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on %s : %w", req.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status %v for request %s", resp.StatusCode, req.URL)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return result, nil
}

// StoreSecret stores secret in keychain item. It uses pass for encrypting the secret.
func (a *ClientApp) StoreSecret(item *keychain.KeychainItem, secret []byte, pass string) error {
	// check pass for old secret
	if oldSecret, err := a.GetSecret(item, pass); err != nil {
		return err
	} else if bytes.Equal(oldSecret, secret) {
		// no changes
		return nil
	}

	env, err := a.enc.Encrypt(secret, []byte(pass))
	if err != nil {
		return fmt.Errorf("error storing secret:%w", err)
	}

	item.StoreSecret(env.Key, env.Data)

	return nil
}

// GetSecret reads secret from keychain item. It uses pass for decrypting secret value.
func (a *ClientApp) GetSecret(item *keychain.KeychainItem, pass string) ([]byte, error) {
	data := item.Data()
	if len(data.Value) == 0 {
		return data.Value, nil
	}
	secret, err := a.enc.Decrypt(&encrypter.Envelope{
		Key:  data.Key,
		Data: data.Value,
	}, []byte(pass))

	if err != nil {
		return nil, fmt.Errorf("decryption failed:%w", err)
	}

	return secret, nil
}
