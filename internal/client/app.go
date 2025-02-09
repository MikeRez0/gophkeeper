package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/keychain"
	"go.uber.org/zap"
)

type ClientApp struct {
	config     *config.ConfigClient
	Log        *zap.Logger
	serverHost string
	token      string
	Keychains  []*keychain.Keychain
}

func NewApp(config *config.ConfigClient, log *zap.Logger) (*ClientApp, error) {
	return &ClientApp{
		config:     config,
		Log:        log,
		Keychains:  make([]*keychain.Keychain, 0, 1),
		serverHost: config.HostString,
	}, nil
}

func (a *ClientApp) Connect(login, password string) error {
	req, err := a.prepareRequest(
		"/api/user/login",
		http.MethodPost,
		map[string]string{"login": login, "password": password})
	if err != nil {
		return fmt.Errorf("error on %s : %w", req.URL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error on %s : %w", req.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status %v for request %s", resp.StatusCode, req.URL)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
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

func (a *ClientApp) prepareRequest(path string, method string, data any) (*http.Request, error) {
	requestStr := a.serverHost + path

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

	return req, nil
}
func (a *ClientApp) FetchKeychainList() error {
	req, err := a.prepareRequest("/api/keychain", http.MethodGet, nil)
	if err != nil {
		return fmt.Errorf("error creating req keychain list : %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error on %s : %w", req.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status %v for request %s", resp.StatusCode, req.URL)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	keychainList := make([]domain.KCData, 0)
	err = json.Unmarshal(data, &keychainList)
	if err != nil {
		return fmt.Errorf("error parsing response %w", err)
	}

	for _, kdata := range keychainList {
		k, err := keychain.NewKeychain(&kdata, a.Log)
		if err != nil {
			return fmt.Errorf("error creating keychain: %w", err)
		}
		a.Keychains = append(a.Keychains, k)
	}

	return nil
}

func (a *ClientApp) SyncKeychain(keychain *keychain.Keychain) error {
	items := make([]*domain.KCItemData, 0)
	for _, i := range keychain.Items {
		if i.IsChanged() {
			items = append(items, i.Data())
		}
	}

	req, err := a.prepareRequest(
		"/api/keychain/"+keychain.Data().ID.String()+"/sync",
		http.MethodPost,
		items,
	)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error on %s : %w", req.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status %v for request %s", resp.StatusCode, req.URL)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	clear(items)

	err = json.Unmarshal(data, &items)
	if err != nil {
		return fmt.Errorf("error reading items:%w", err)
	}

	for _, i := range items {
		keychain.ApplyItemFromData(i)
	}

	return nil
}
