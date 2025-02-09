package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type KeychainID uuid.UUID

type KCData struct {
	Name    string     `json:"name"`
	ID      KeychainID `json:"id"`
	OwnerID UserID     `json:"owner_id"`
}

type KCItemType byte

const (
	KCItemTypePassword   KCItemType = iota + 1
	KCItemTypeString     KCItemType = iota + 1
	KCItemTypeBinary     KCItemType = iota + 1
	KCItemTypeCardNumber KCItemType = iota + 1
)

type KeychainItemID uuid.UUID

const (
	KCMetaKeyComment = "Comment"
	KCMetaKeyLogin   = "Login"
	KCMetaKeySite    = "Site"
	KCMetaKeyIssuer  = "Issuer"
	KCMetaKeyOwner   = "Owner"
	KCMetaKeyValidTo = "ValidTo"
)

type KeychainItemMeta map[string]string

type KCItemData struct {
	Label      string           `json:"label"`
	ClientTime time.Time        `json:"client_time"`
	ServerTime time.Time        `json:"server_time"`
	MetaData   KeychainItemMeta `json:"meta"`
	Value      []byte           `json:"value"`
	Key        []byte           `json:"key"`
	KeyChainID KeychainID       `json:"keychain_id"`
	ID         KeychainItemID   `json:"id"`
	ItemType   KCItemType       `json:"type"`
}

func (md KeychainItemMeta) String() string {
	str := ""
	for k, v := range md {
		str += fmt.Sprintf("%s: %s; ", k, v)
	}
	return str
}

func (k KeychainID) Value() (driver.Value, error) {
	return uuid.UUID(k).String(), nil
}

func (k KeychainID) String() string {
	return uuid.UUID(k).String()
}

func (k *KeychainID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", uuid.UUID(*k).String())), nil
}

func (k *KeychainID) UnmarshalJSON(b []byte) error {
	u, err := uuid.Parse(strings.Trim(string(b), "\""))
	if err != nil {
		return errors.New("could not parse UUID")
	}
	*k = KeychainID(u)
	return nil
}

func (k *KeychainItemID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", uuid.UUID(*k).String())), nil
}

func (k *KeychainItemID) UnmarshalJSON(b []byte) error {
	u, err := uuid.Parse(strings.Trim(string(b), "\""))
	if err != nil {
		return errors.New("could not parse UUID")
	}
	*k = KeychainItemID(u)
	return nil
}

func (k KeychainItemID) Value() (driver.Value, error) {
	return uuid.UUID(k).String(), nil
}

func (k KeychainItemID) String() string {
	return uuid.UUID(k).String()
}
