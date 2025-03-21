// Package domain contaions model data types.
// # Base objects:
//   - Keychain - storage for secret items, owned by user.
//   - Keychain Item - object in keychain that contains label, meta data and secret.
//
// Item consists of:
//   - Label
//   - Secret value
//   - Additional (meta) data for item: comment, etc. All items have at least comment.
//
// Encrypting method:
// Secret value stored in encrypted format (with user password).
//  1. Secret value encrypted by random 64-bit key.
//  2. Random 64-bit key encrypted with user's keychain password.
package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type KeychainID uuid.UUID // Keychain header ID

// KCData stores information about keychain header.
type KCData struct {
	Name    string     `json:"name"`
	ID      KeychainID `json:"id"`
	OwnerID UserID     `json:"-"`
}

type KCItemType byte // Keychain item type controls meta data for item.

const (
	KCItemTypePassword   KCItemType = iota + 1 // password for site. Meta data: login, site
	KCItemTypeString     KCItemType = iota + 1 // just secret string
	KCItemTypeBinary     KCItemType = iota + 1 // binary secret data (file). Meta data: filename
	KCItemTypeCardNumber KCItemType = iota + 1 // credit card data. Meta data: owner, issuer, validTo
)

type KeychainItemID uuid.UUID // Keychain item ID

const (
	KCMetaKeyComment  = "Comment"
	KCMetaKeyLogin    = "Login"
	KCMetaKeySite     = "Site"
	KCMetaKeyIssuer   = "Issuer"
	KCMetaKeyOwner    = "Owner"
	KCMetaKeyValidTo  = "ValidTo"
	KCMetaKeyFilename = "Filename"
)

type KeychainItemMeta map[string]string // map to store item's meta data.

// KCItemData stores information about item.
type KCItemData struct {
	Label      string           `json:"label"`       // Label
	ClientTime time.Time        `json:"client_time"` // Client update time (key field for synchronisation)
	ServerTime time.Time        `json:"server_time"` // Server update time
	MetaData   KeychainItemMeta `json:"meta"`        // Meta data for item
	Value      []byte           `json:"value"`       // Encrypted secret value
	Key        []byte           `json:"key"`         // Encrypted key for secret value
	KeyChainID KeychainID       `json:"keychain_id"` // Keychain ID
	ID         KeychainItemID   `json:"id"`          // Item ID
	ItemType   KCItemType       `json:"type"`        // Item type
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

func (k *KeychainID) Scan(value interface{}) error {
	u := uuid.UUID(*k)
	err := u.Scan(value)
	if err != nil {
		return fmt.Errorf("scan error: %w", err)
	}
	*k = KeychainID(u)
	return nil
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

func (k *KeychainItemID) Scan(value interface{}) error {
	u := uuid.UUID(*k)
	err := u.Scan(value)
	if err != nil {
		return fmt.Errorf("scan error: %w", err)
	}
	*k = KeychainItemID(u)
	return nil
}

func (k KCItemType) String() string {
	switch k {
	case KCItemTypeBinary:
		return "Binary data"
	case KCItemTypeCardNumber:
		return "Card number"
	case KCItemTypePassword:
		return "Login/pass"
	case KCItemTypeString:
		return "Secret text"
	default:
		return "Unknown item"
	}
}

// KeyChainTypes - string list of types of keychain items.
func KeyChainTypes() []string {
	result := make([]string, 0, 5)
	for i := 1; i < 5; i++ {
		result = append(result, KCItemType(i).String())
	}
	return result
}
