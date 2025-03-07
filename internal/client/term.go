package client

import (
	"context"
	"fmt"

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
}

func NewCommandExecutor(app *app.ClientApp) *CommandExecutor {
	return &CommandExecutor{app: app, OfflineMode: false}
}

func (c *CommandExecutor) preCommand(ctx context.Context) error {
	_ = c.sync(ctx)

	return nil
}

func (c *CommandExecutor) sync(ctx context.Context) error {
	if c.OfflineMode {
		return nil
	}

	if c.Token != "" {
		c.app.SetToken(c.Token)
	} else {
		if c.Login == "" {
			c.Login = c.inputString("Login", "", false)
		}
		if c.Password == "" {
			c.Password = c.inputString("Password", "", true)
		}
		fmt.Print("Connecting ... ")
		err := c.app.Connect(ctx, c.Login, c.Password)
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
			item.MetaDataSetItem(k, c.inputString(k, v, false))
		}

		err := c.app.StoreSecret(item,
			[]byte(c.inputString("SECRET", string(secret), true)),
			c.KeychainPass)
		if err != nil {
			return fmt.Errorf("store secret error: %w", err)
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
			fmt.Println(string(secret))
			return nil
		}

		info = append(info,
			infoS{"Type", item.Data().ItemType.String()},
			infoS{"Label", item.Label()})

		// Meta data
		for k, v := range item.MetaData() {
			info = append(info, infoS{k, v})
		}

		info = append(info, infoS{"SECRET", string(secret)})
	} else {
		return fmt.Errorf("get secret error: %w", err)
	}

	return writeTab(info)
}
