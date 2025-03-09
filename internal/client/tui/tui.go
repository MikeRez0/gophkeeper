package tui

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/client/app"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/keychain"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

const (
	cPageKeychain  = "keychain"
	cPageItemsList = "items"
	cPageLogin     = "login"
)

type UIController struct { //nolint:govet // for comfortable work
	log *zap.Logger
	app *app.ClientApp

	login            string
	password         string
	keychainPassword string
	secretValue      []byte

	keychainID     domain.KeychainID
	keychainItemID domain.KeychainItemID
	keychainItem   *keychain.KeychainItem

	uiapp        *tview.Application
	pages        *tview.Pages
	keychainList *tview.List
	itemsList    *tview.List
	itemForm     *tview.Form
	passForm     *tview.Form
	loginForm    *tview.Form
	logView      *tview.TextView
}

func NewUIController(app *app.ClientApp, log *zap.Logger) (*UIController, error) {
	c := &UIController{
		app: app,
		log: log,
	}

	return c, nil
}

func (c *UIController) buildUI() {
	c.uiapp = tview.NewApplication()
	c.pages = tview.NewPages()

	c.keychainList = tview.NewList()
	c.keychainList.SetBorder(true)
	c.keychainList.SetSecondaryTextColor(tcell.ColorDarkKhaki)

	c.itemsList = tview.NewList()
	c.itemsList.SetBorder(true)

	c.itemForm = tview.NewForm()
	c.itemForm.SetBorder(true)

	c.passForm = tview.NewForm()
	c.passForm.SetBorder(true)

	c.loginForm = tview.NewForm()
	c.loginForm.SetBorder(true)

	c.logView = tview.NewTextView()
	c.logView.SetBorder(true)

	pLogin := tview.NewGrid().
		AddItem(tview.NewBox(), 0, 0, 12, 12, 0, 0, false).
		AddItem(c.loginForm, 1, 1, 9, 9, 10, 15, true).
		AddItem(c.logView, 12, 0, 1, 12, 1, 30, true)
	pLogin.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			c.uiapp.Stop()
		}
		return event
	})

	pKeychain := tview.NewGrid().
		AddItem(c.keychainList, 0, 0, 9, 2, 8, 10, true).
		AddItem(c.passForm, 0, 2, 9, 10, 8, 15, true).
		AddItem(c.logView, 9, 0, 2, 12, 2, 0, true).
		AddItem(tview.NewTextView().SetText("(Q) - quit (S) - sync"),
			11, 0, 1, 12, 0, 0, false)

	pKeychain.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'q':
			if c.keychainList.HasFocus() {
				c.uiapp.Stop()
			}
		case event.Rune() == 's':
			if c.itemsList.HasFocus() || c.keychainList.HasFocus() {
				c.sync()
			}
		}
		return event
	})

	pItems := tview.NewGrid().
		AddItem(c.keychainList, 0, 0, 9, 2, 8, 10, false).
		AddItem(c.itemsList, 0, 2, 9, 4, 8, 15, true).
		AddItem(c.itemForm, 0, 6, 9, 6, 8, 30, true).
		AddItem(c.logView, 9, 0, 2, 12, 2, 0, true).
		AddItem(tview.NewTextView().SetText("(A) - add item (Q) - quit (S) - sync"),
			11, 0, 1, 12, 0, 0, false)

	pItems.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'q':
			if c.itemsList.HasFocus() {
				c.uiapp.Stop()
			}
		case event.Rune() == 'a':
			if c.itemsList.HasFocus() {
				c.keychainItemID = domain.KeychainItemID(uuid.Nil)
				c.showItemForm()
				return nil
			}
		case event.Rune() == 's':
			if c.itemsList.HasFocus() || c.keychainList.HasFocus() {
				c.sync()
			}
		case event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEscape:
			if c.itemsList.HasFocus() {
				c.keychainPassword = ""
				c.showKeychainList()
			}
		}

		return event
	})

	c.pages.AddPage(cPageLogin, pLogin, true, false)
	c.pages.AddPage(cPageKeychain, pKeychain, true, false)
	c.pages.AddPage(cPageItemsList, pItems, true, false)
}

func (c *UIController) scheduleSync(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.sync()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *UIController) sync() {
	ctx := context.Background()

	syncChan, err := c.app.RunSync(ctx)
	if err != nil {
		c.writeLog("run sync error", err)
	} else {
		c.writeLog("sync started", nil)

		go func() {
			syncStatus := <-syncChan
			c.uiapp.QueueUpdateDraw(func() {
				if syncStatus.Error != nil {
					c.writeLog("sync error", err)
				} else {
					c.writeLog("sync done "+
						fmt.Sprintf("uploaded: %d, downloaded: %d, force updated: %d",
							syncStatus.SyncTaskResult.Uploaded,
							syncStatus.SyncTaskResult.Downloaded,
							syncStatus.SyncTaskResult.UpdatedForNewerVersion),
						nil)
					c.update()
				}
			})
		}()
	}
}

func (c *UIController) update() {
	// render keychain list
	c.keychainList.Clear()
	c.keychainList.ShowSecondaryText(false)

	keychains, err := c.app.Service.KeychainList(context.Background(), 0)
	if err != nil {
		c.writeLog("list keychain error", err)
		return
	}

	selected := 0
	for i, k := range keychains {
		c.keychainList.AddItem(
			fmt.Sprintf("%v:  %s", i, k.Name),
			"", rune(0), func() {
				c.keychainID = k.ID
				c.requestPass()
			})
		if k.ID == c.keychainID {
			selected = i
		}
	}

	c.keychainList.SetCurrentItem(selected)

	// render items list
	c.itemsList.Clear()
	ctx := context.Background()

	list, err := c.app.Service.KeychainGetItemsSince(ctx, 0, c.keychainID, time.Time{})
	if err != nil && !errors.Is(err, domain.ErrDataNotFound) {
		c.writeLog("read list error", err)
		return
	}

	selected = 0
	for i, item := range list {
		c.itemsList.AddItem(
			fmt.Sprintf("%v: %s", i, item.Label),
			fmt.Sprintf("[%s] %s", item.ItemType, item.MetaData[domain.KCMetaKeyComment]),
			rune(0), func() {
				c.keychainItemID = item.ID
				c.showItemForm()
			})
		if item.ID == c.keychainItemID {
			selected = i
		}
	}

	c.itemsList.SetCurrentItem(selected)
}

func (c *UIController) showKeychainList() {
	c.keychainID = domain.KeychainID{}
	c.itemsList.Clear()

	c.update()

	c.pages.SwitchToPage(cPageKeychain)
	c.uiapp.SetFocus(c.keychainList)
}

func (c *UIController) showItems() {
	// clear item detail
	c.secretValue = nil
	c.itemForm.Clear(true)
	c.resetKeychainItem()

	c.update()

	c.uiapp.SetFocus(c.itemsList)
}

func (c *UIController) setKeychainItem(kid domain.KeychainItemID) error {
	if c.keychainItem == nil {
		if kid == domain.KeychainItemID(uuid.Nil) {
			// factory new keychain item
			c.keychainItem = keychain.NewKeychainItem(c.keychainID, domain.KCItemTypePassword)
			c.keychainItemID = c.keychainItem.Data().ID
			return nil
		} else {
			// load item from local store
			data, err := c.app.Service.KeychainGetItem(context.Background(), 0,
				c.keychainID, c.keychainItemID)
			if err != nil {
				c.writeLog("error reading item", err)
				return err
			}
			c.keychainItem = keychain.NewKeychainItemFromData(data)
			c.keychainItemID = c.keychainItem.Data().ID
			return nil
		}
	} else {
		if kid != c.keychainItem.Data().ID {
			c.keychainItem = nil
			return c.setKeychainItem(kid)
		}
		return nil
	}
}
func (c *UIController) resetKeychainItem() {
	c.keychainItem = nil
	c.keychainItemID = domain.KeychainItemID(uuid.Nil)
}

func (c *UIController) showItemForm() {
	c.itemForm.Clear(true)

	err := c.setKeychainItem(c.keychainItemID)
	if err != nil {
		c.writeLog("error loading item", err)
		return
	}

	item := c.keychainItem

	if secret, err := c.app.GetSecret(item, c.keychainPassword); err == nil {
		c.itemForm.AddInputField("Label",
			item.Label(), 40, nil,
			func(text string) {
				item.SetLabel(text)
			})

		c.itemForm.AddDropDown("Type",
			domain.KeyChainTypes(),
			int(item.Data().ItemType)-1,
			func(option string, optionIndex int) {
				if domain.KCItemType(optionIndex+1) != item.Data().ItemType {
					item.SetType(domain.KCItemType(optionIndex + 1))
					c.showItemForm()
				}
			})

		c.renderMetaData(item)

		if item.Data().ItemType != domain.KCItemTypeBinary {
			c.secretValue = secret
			c.itemForm.AddInputField("Secret",
				string(c.secretValue),
				40,
				nil, func(text string) {
					c.secretValue = []byte(text)
				})
		} else {
			c.itemForm.AddTextView("Secret", "Work with binary secret with CLI", 50, 1, false, false)
		}
		c.itemForm.AddButton("OK", func() {
			if item.Data().ItemType != domain.KCItemTypeBinary {
				_ = c.app.StoreSecret(item, c.secretValue, c.keychainPassword)
			}

			_, _, err := c.app.Service.KeychainSaveItem(context.Background(),
				c.app.UserID, item.Data())
			if err != nil {
				c.writeLog("store items error", err)
				return
			}
			c.showItems()
		})
		c.itemForm.AddButton("Cancel", func() {
			c.showItems()
		})
	} else {
		c.writeLog("error reading secret", err)
		c.itemForm.AddTextView("Error", "Wrong pass key", 30, 1, false, false)
		c.itemForm.AddButton("OK", func() {
			c.showItems()
		})
	}

	c.uiapp.SetFocus(c.itemForm)
}

func (c *UIController) renderMetaData(keychainItem *keychain.KeychainItem) {
	for k, v := range keychainItem.MetaData() {
		c.itemForm.AddInputField(k,
			v, 50, nil,
			func(text string) {
				keychainItem.MetaDataSetItem(k, text)
			})
	}
}

func (c *UIController) requestPass() {
	c.passForm.Clear(true)

	c.passForm.AddPasswordField("Pass",
		c.keychainPassword, 30, '*',
		func(text string) {
			c.keychainPassword = text
		})
	c.passForm.AddButton("OK", func() {
		c.passForm.Clear(true)
		c.showItems()
		c.pages.SwitchToPage("items")
	})
	c.passForm.AddButton("Cancel", func() {
		c.passForm.Clear(true)
		c.keychainPassword = ""
		c.showKeychainList()
	})

	c.uiapp.SetFocus(c.passForm)
}

func (c *UIController) showLoginForm() {
	c.loginForm.Clear(true)

	c.loginForm.AddInputField(
		"Login",
		c.login, 20,
		nil,
		func(text string) {
			c.login = text
		},
	)

	c.loginForm.AddPasswordField(
		"Password",
		c.password, 20,
		'*',
		func(text string) {
			c.password = text
		},
	)

	c.loginForm.AddButton("Login",
		func() {
			ctx := context.Background()
			err := c.app.Connect(ctx, c.login, c.password)
			if err != nil {
				c.writeLog("connection error", err)
				return
			}

			if c.app.SyncInterval != time.Duration(0) {
				c.scheduleSync(ctx, c.app.SyncInterval)
			}

			c.showKeychainList()
		})

	c.loginForm.AddButton("Register",
		func() {
			ctx := context.Background()
			err := c.app.RegisterUser(ctx, c.login, c.password)
			if err != nil {
				c.writeLog("registration error", err)
				return
			}

			_, err = c.app.Service.KeychainSave(ctx, c.app.UserID, &domain.KCData{
				Name:    "My keychain",
				OwnerID: c.app.UserID,
				ID:      domain.KeychainID(uuid.New()),
			})
			if err != nil {
				c.writeLog("creation keychain error", err)
				return
			}

			if c.app.SyncInterval != time.Duration(0) {
				c.scheduleSync(ctx, c.app.SyncInterval)
			}

			c.showKeychainList()
		})

	c.loginForm.AddButton("Offline",
		func() {
			l, err := c.app.Service.KeychainList(context.Background(), c.app.UserID)
			if err != nil {
				c.writeLog("local keychain read error", err)
				return
			}
			if len(l) == 0 {
				c.writeLog("local keychain is empty", nil)
				return
			}
			c.showKeychainList()
		})

	c.uiapp.SetFocus(c.loginForm)
}

func (c *UIController) writeLog(message string, err error) {
	if err != nil {
		// 2006.01.02 15:04:05
		_, _ = fmt.Fprintf(c.logView, "%s %s: %v\n", time.Now().Format(time.TimeOnly), message, err)
	} else {
		_, _ = fmt.Fprintln(c.logView, time.Now().Format(time.TimeOnly)+" "+message)
	}
	c.logView.ScrollToEnd()
}

func (c *UIController) Run() error {
	c.buildUI()
	c.pages.SwitchToPage(cPageLogin)
	c.showLoginForm()

	if err := c.uiapp.SetRoot(c.pages, true).EnableMouse(false).Run(); err != nil {
		return fmt.Errorf("error in UI app: %w", err)
	}
	return nil
}
