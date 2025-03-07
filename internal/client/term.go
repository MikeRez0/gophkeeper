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

type CommandExecutor struct {
	app *app.ClientApp

	keychainID domain.KeychainID

	OfflineMode  bool
	Login        string
	Password     string
	Token        string
	KeychainPass string
	Filename     string
}

func NewCommandExecutor(app *app.ClientApp) *CommandExecutor {
	return &CommandExecutor{app: app, OfflineMode: false}
}

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
			return err
		} else {
			fmt.Println("done")
		}
	}

	fmt.Print("Start synchronisation... ")
	err := c.app.SyncKeychains(ctx)
	if err != nil {
		fmt.Println("failed")
		c.app.Log.Error("sync error", zap.Error(err))
		return err
	} else {
		fmt.Println("done")
	}
	return nil
}

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

func (c *CommandExecutor) KeychainList(ctx context.Context) error {
	err := c.preCommand(ctx)
	if err != nil {
		return err
	}

	list, err := c.app.Service.KeychainList(ctx, c.app.UserID)
	if err != nil {
		return fmt.Errorf("read keychain list error: %w", err)
	}

	return writeKeychainList(list)
}

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

	return writeItemsList(list)
}

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

	return nil
}

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

	return writeTab(info)
}
