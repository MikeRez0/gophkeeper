package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/keychain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/MikeRez0/gophkeeper/internal/core/utils/encrypter"
	"go.uber.org/zap"
)

type ClientApp struct {
	config   *config.ConfigClient
	enc      *encrypter.Encrypter
	dec      *encrypter.Decrypter
	Log      *zap.Logger
	token    string
	Service  port.IKeychainDataService
	SyncTime time.Time
	UserID   domain.UserID
}

const cKeySize = 32

func NewApp(config *config.ConfigClient, service port.IKeychainDataService, log *zap.Logger) (*ClientApp, error) {
	enc, err := encrypter.NewEncrypter(log, cKeySize)
	if err != nil {
		return nil, fmt.Errorf("error creating encrypter: %w", err)
	}
	dec, err := encrypter.NewDecrypter(log, cKeySize)
	if err != nil {
		return nil, fmt.Errorf("error creating decrypter: %w", err)
	}

	return &ClientApp{
		config:   config,
		Log:      log,
		enc:      enc,
		dec:      dec,
		Service:  service,
		SyncTime: time.Time{},
		UserID:   domain.UserID(0),
	}, nil
}

func (a *ClientApp) SetToken(token string) {
	a.token = token
}

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

	resp, err := http.DefaultClient.Do(req)
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

func (a *ClientApp) GetSecret(item *keychain.KeychainItem, pass string) ([]byte, error) {
	data := item.Data()
	if len(data.Value) == 0 {
		return data.Value, nil
	}
	secret, err := a.dec.Decrypt(&encrypter.Envelope{
		Key:  data.Key,
		Data: data.Value,
	}, []byte(pass))

	if err != nil {
		return nil, fmt.Errorf("decryption failed:%w", err)
	}

	return secret, nil
}
