package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/MikeRez0/gophkeeper/internal/client/app"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/keychain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CommandExecutor core object for executing commands.
type CommandExecutor struct {
	app *app.ClientApp // client app core

	Login        string            // login from command flags
	Password     string            // password from command flags
	Token        string            // token value from command flags
	KeychainPass string            // keychain password from command flags
	Filename     string            // file name for binary secrets from command flags
	keychainID   domain.KeychainID // current keychain ID
	OfflineMode  bool              // is offline mode from command flags
}

// NewCommandExecutor creates new CommandExecutor.
func NewCommandExecutor(a *app.ClientApp) *CommandExecutor {
	return &CommandExecutor{app: a, OfflineMode: false}
}

// preCommand runs sync job if needed.
func (c *CommandExecutor) preCommand(ctx context.Context) error {
	if !c.OfflineMode {
		_ = c.sync(ctx)
	}

	return nil
}

func (c *CommandExecutor) sync(ctx context.Context) error {
	if c.Token != "" {
		c.app.SetToken(c.Token)
	} else {
		fmt.Print("Connecting ... ")
		err := c.app.Connect(ctx, c.requestLogin(), c.requestPassword())
		if err != nil {
			fmt.Println("failed")
			c.app.Log.Error("connection error", zap.Error(err))
			return fmt.Errorf("connection error: %w", err)
		} else {
			fmt.Println("done")
		}
	}

	fmt.Print("Start synchronisation... ")
	_, err := c.app.SyncKeychains(ctx)
	if err != nil {
		fmt.Println("failed")
		c.app.Log.Error("sync error", zap.Error(err))
		return fmt.Errorf("sync error: %w", err)
	} else {
		fmt.Println("done")
	}
	return nil
}

// Register runs registration new user on server.
// Creates new local keychain if it's not exist.
func (c *CommandExecutor) Register(ctx context.Context) error {
	list, err := c.app.Service.KeychainList(ctx, c.app.UserID)
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		return fmt.Errorf("read keychain list error: %w", err)
	}

	if errors.Is(err, domain.ErrDataNotFound) || len(list) == 0 {
		_, err := c.app.Service.KeychainSave(ctx, c.app.UserID,
			&domain.KCData{
				Name:    "My keychain",
				ID:      domain.KeychainID(uuid.New()),
				OwnerID: c.app.UserID,
			})
		if err != nil {
			return fmt.Errorf("error creating local keychain: %w", err)
		}
	}

	err = c.app.RegisterUser(ctx, c.requestLogin(), c.requestPassword())
	if err != nil {
		return fmt.Errorf("registration error: %w", err)
	}

	return c.sync(ctx)
}

// KeychainAdd creates new local keychain with given name.
func (c *CommandExecutor) KeychainAdd(ctx context.Context, name string) error {
	err := c.preCommand(ctx)
	if err != nil {
		return err
	}

	k, err := c.app.Service.KeychainSave(ctx, c.app.UserID, &domain.KCData{
		Name:    name,
		ID:      domain.KeychainID(uuid.New()),
		OwnerID: c.app.UserID,
	})
	if err != nil {
		return fmt.Errorf("save keychain error: %w", err)
	}

	fmt.Println("Keychain saved. ID: " + k.ID.String())

	return nil
}

// KeychainList writes keychain list.
func (c *CommandExecutor) KeychainList(ctx context.Context) error {
	err := c.preCommand(ctx)
	if err != nil {
		return err
	}

	list, err := c.app.Service.KeychainList(ctx, c.app.UserID)
	if err != nil {
		return fmt.Errorf("read keychain list error: %w", err)
	}

	err = writeKeychainList(list)
	if err != nil {
		return err
	}

	return nil
}

// ItemList writes keychain items list.
// It uses map [flags] for filtering items by label and metadata values.
func (c *CommandExecutor) ItemList(ctx context.Context, flags map[string]string) error {
	err := c.preCommand(ctx)
	if err != nil {
		return err
	}

	if keychainID, err := c.findKeychainID(ctx, flags); err == nil {
		c.keychainID = keychainID
	} else {
		return err
	}

	list, err := c.queryKeychainItem(ctx, c.keychainID, flags)
	if err != nil {
		return fmt.Errorf("read local items error: %w", err)
	}

	err = writeItemsList(list)
	if err != nil {
		return err
	}

	return nil
}

// ItemStore saves keychain item.
// It uses map [flags] to find item to update by label and metadata values .
// If item not found it creates new one.
// Algorithm:
// 1. Determine keychain
// 2. Request keychain password if needed
// 3. Query items from local storage (select by user if found more than one)
// 4. Create items if not found
// 5. Read secret (check that password is correct)
// 6. Input params of item (type, label, metadata)
// 7. Input secret value
// 8. Save item
//
// For binary type read secret value from file.
func (c *CommandExecutor) ItemStore(ctx context.Context, flags map[string]string) error {
	err := c.preCommand(ctx)
	if err != nil {
		return err
	}

	if keychainID, err := c.findKeychainID(ctx, flags); err == nil {
		c.keychainID = keychainID
	} else {
		return err
	}

	c.requestKeychainPass()

	var item *keychain.KeychainItem
	if len(flags) != 0 {
		itemData, err := c.findKeychainItem(ctx, c.keychainID, flags)
		if err != nil {
			return fmt.Errorf("item not found: %w", err)
		}
		item = keychain.NewKeychainItemFromData(itemData)
	} else {
		item = keychain.NewKeychainItem(c.keychainID, domain.KCItemTypePassword)
	}

	if secret, err := c.app.GetSecret(item, c.KeychainPass); err == nil {
		// Keychain item type
		g := "Select item type: \n"
		tl := domain.KeyChainTypes()
		for i, t := range tl {
			g += fmt.Sprintf("%d - %s\n", i+1, t)
		}
		item.SetType(domain.KCItemType(
			c.inputNumber(g, int(item.Data().ItemType), len(tl))))

		// Label
		item.SetLabel(c.inputString("Label", item.Label(), false))

		// Meta data
		for k, v := range item.MetaData() {
			if k == domain.KCMetaKeyFilename {
				continue
			}
			item.MetaDataSetItem(k, c.inputString(k, v, false))
		}

		if item.Data().ItemType == domain.KCItemTypeBinary {
			item.MetaDataSetItem(domain.KCMetaKeyFilename, path.Base(c.requestFilename()))

			secret, err := os.ReadFile(c.Filename)
			if err != nil {
				return fmt.Errorf("read file error: %w", err)
			}
			err = c.app.StoreSecret(item, secret, c.KeychainPass)
			if err != nil {
				return fmt.Errorf("store secret error: %w", err)
			}
		} else {
			err := c.app.StoreSecret(item,
				[]byte(c.inputString("SECRET", string(secret), true)),
				c.KeychainPass)
			if err != nil {
				return fmt.Errorf("store secret error: %w", err)
			}
		}
	} else {
		return fmt.Errorf("get secret error: %w", err)
	}

	_, _, err = c.app.Service.KeychainSaveItem(ctx,
		c.app.UserID,
		item.Data())

	if err != nil {
		return fmt.Errorf("store item to local keychain error: %w", err)
	}

	fmt.Println("Item saved")

	return c.preCommand(ctx)
}

// ItemShow write to stdout keychain item.
// It uses map [flags] to find item by label and metadata values .
// Algorithm:
// 1. Determine keychain
// 2. Request keychain password if needed
// 3. Query items from local storage (select by user if found more than one)
// 4. Print item
//
// For binary type write secret value to file.
func (c *CommandExecutor) ItemShow(ctx context.Context, flags map[string]string, onlySecret bool) error {
	if err := c.preCommand(ctx); err != nil {
		return err
	}

	if keychainID, err := c.findKeychainID(ctx, flags); err == nil {
		c.keychainID = keychainID
	} else {
		return err
	}

	c.requestKeychainPass()

	var item *keychain.KeychainItem
	if d, err := c.findKeychainItem(ctx, c.keychainID, flags); err == nil {
		item = keychain.NewKeychainItemFromData(d)
	} else {
		return fmt.Errorf("search item error: %w", err)
	}

	info := make([]infoS, 0)

	// Secret
	if secret, err := c.app.GetSecret(item, c.KeychainPass); err == nil {
		if onlySecret {
			if item.Data().ItemType != domain.KCItemTypeBinary {
				fmt.Println(string(secret))
				return nil
			} else {
				if err := binary.Write(os.Stdout, binary.BigEndian, secret); err != nil {
					return fmt.Errorf("error writing binary secret: %w", err)
				}
				return nil
			}
		}

		info = append(info,
			infoS{"Type", item.Data().ItemType.String()},
			infoS{"Label", item.Label()})

		// Meta data
		for k, v := range item.MetaData() {
			info = append(info, infoS{k, v})
		}

		if item.Data().ItemType == domain.KCItemTypeBinary {
			if c.Filename == "" {
				c.Filename = item.MetaDataItem(domain.KCMetaKeyFilename)
			}
			f, err := os.Create(c.requestFilename())
			if err != nil {
				return fmt.Errorf("create file error: %w", err)
			}
			defer func() { _ = f.Close() }()

			_, err = f.Write(secret)
			if err != nil {
				return fmt.Errorf("write file error: %w", err)
			}

			info = append(info, infoS{"SECRET", fmt.Sprintf("  *** writed to %s ***\n", c.Filename)})
		} else {
			info = append(info, infoS{"SECRET", string(secret)})
		}
	} else {
		return fmt.Errorf("get secret error: %w", err)
	}

	err := writeTab(info)
	if err != nil {
		return err
	}

	return nil
}
