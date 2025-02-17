package tui

import (
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/client/app"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/keychain"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

const (
	cPageKeychain  = "keychain"
	cPageItemsList = "items"
	cPageLogin     = "login"
)

type UIController struct {
	log *zap.Logger
	app *app.ClientApp

	uiapp        *tview.Application
	pages        *tview.Pages
	keychainList *tview.List
	itemsList    *tview.List
	itemForm     *tview.Form
	passForm     *tview.Form
	loginForm    *tview.Form
	logView      *tview.TextView

	login    string
	password string

	keychain     *keychain.Keychain
	keychainItem *keychain.KeychainItem
	secretValue  []byte
}

func NewUIController(app *app.ClientApp, log *zap.Logger) (*UIController, error) {
	c := &UIController{
		app: app,
		log: log,
	}

	return c, nil
}

func (c *UIController) refresh() {
	switch {
	case c.keychain == nil: // keychain list is active
		c.uiapp.SetFocus(c.keychainList)
	case c.keychainItem == nil:
		c.uiapp.SetFocus(c.itemsList)
	case c.keychainItem != nil:
		c.uiapp.SetFocus(c.itemForm)
	}
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
		AddItem(c.loginForm, 3, 3, 6, 3, 10, 15, true).
		AddItem(c.logView, 12, 0, 1, 12, 1, 30, true)

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
				err := c.app.SyncKeychain(c.keychain)
				if err != nil {
					c.log.Error("sync error", zap.Error(err))
					c.writeLog("sync error", err)
				} else {
					c.writeLog("successfull synced", nil)
				}
				if c.keychain != nil {
					c.showItems()
				}
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
				c.keychainItem = c.keychain.NewItem(domain.KCItemTypePassword)
				c.showItemForm()
			}
		case event.Rune() == 's':
			if c.itemsList.HasFocus() || c.keychainList.HasFocus() {
				err := c.app.SyncKeychain(c.keychain)
				if err != nil {
					c.writeLog("sync error", err)
				} else {
					c.writeLog("successfull synced", nil)
				}
				if c.keychain != nil {
					c.showItems()
				}
			}
		case event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyEscape:
			if c.itemsList.HasFocus() {
				c.keychain.Pass = ""
				c.showKeychainList()
			}
		}

		return event
	})

	c.pages.AddPage(cPageLogin, pLogin, true, false)
	c.pages.AddPage(cPageKeychain, pKeychain, true, false)
	c.pages.AddPage(cPageItemsList, pItems, true, false)
}

func (c *UIController) showKeychainList() {
	c.keychain = nil
	c.itemsList.Clear()

	c.keychainList.Clear()
	c.keychainList.ShowSecondaryText(false)

	selected := 0
	for i, k := range c.app.Keychains {
		c.keychainList.AddItem(
			fmt.Sprintf("%v:  %s", i, k.Data().Name),
			"", rune(0), func() {
				c.keychain = k
				c.requestPass()
			})
		if k == c.keychain {
			selected = i
		}
	}

	c.keychainList.SetCurrentItem(selected)
	c.pages.SwitchToPage(cPageKeychain)
	c.refresh()
	c.uiapp.SetFocus(c.keychainList)
}

func (c *UIController) showItems() {
	c.itemsList.Clear()

	for i, item := range c.keychain.Items {
		status := ""
		if item.IsChanged() {
			status = "*"
		}
		c.itemsList.AddItem(
			fmt.Sprintf("%v: %s %s", i, item.Label(), status),
			fmt.Sprintf("[%s] %s", item.Data().ItemType, item.MetaDataItem(domain.KCMetaKeyComment)),
			rune(0), func() {
				c.keychainItem = item
				c.showItemForm()
			})
	}

	c.refresh()
	c.uiapp.SetFocus(c.itemsList)
}

func (c *UIController) showItemForm() {
	c.itemForm.Clear(true)

	item := c.keychainItem

	if secret, err := c.keychain.GetSecret(c.keychainItem); err == nil {
		c.itemForm.AddInputField("Label",
			item.Label(), 40, nil,
			func(text string) {
				item.SetLabel(text)
			})

		c.itemForm.AddDropDown("Type",
			domain.KeyChainTypes(),
			int(item.Data().ItemType)-1,
			func(option string, optionIndex int) {
				if domain.KCItemType(optionIndex+1) != c.keychainItem.Data().ItemType {
					c.keychainItem.SetType(domain.KCItemType(optionIndex + 1))
					c.showItemForm()
				}
			})

		c.renderMetaData()
		c.itemForm.AddButton("OK", func() {
			_ = c.keychain.StoreSecret(c.keychainItem, c.secretValue)
			c.secretValue = nil
			c.itemForm.Clear(true)
			c.keychainItem = nil
			c.showItems()
		})

		c.secretValue = secret
		c.itemForm.AddInputField("Secret",
			string(c.secretValue),
			40,
			nil, func(text string) {
				c.secretValue = []byte(text)
			})
	} else {
		c.writeLog("error reading secret", err)
		c.itemForm.AddTextView("Error", "Wrong pass key", 30, 1, false, false)
		c.itemForm.AddButton("OK", func() {
			c.secretValue = nil
			c.itemForm.Clear(true)
			c.keychainItem = nil
			c.showItems()
		})
	}

	c.uiapp.SetFocus(c.itemForm)
}

func (c *UIController) renderMetaData() {
	for k, v := range c.keychainItem.MetaData() {
		c.itemForm.AddInputField(k,
			v, 50, nil,
			func(text string) {
				c.keychainItem.MetaDataSetItem(k, text)
			})
	}
}

func (c *UIController) requestPass() {
	c.passForm.Clear(true)

	c.passForm.AddPasswordField("Pass",
		c.keychain.Pass, 30, '*',
		func(text string) {
			c.keychain.Pass = text
		})
	c.passForm.AddButton("OK", func() {
		c.passForm.Clear(true)
		c.showItems()
		c.pages.SwitchToPage("items")
	})
	c.passForm.AddButton("Cancel", func() {
		c.passForm.Clear(true)
		c.keychain.Pass = ""
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
			err := c.app.Connect(c.login, c.password)
			if err != nil {
				c.writeLog("connection error", err)
				return
			}
			err = c.app.FetchKeychainList()
			if err != nil {
				c.writeLog("fetch keychain list error", err)
				return
			}
			err = c.app.SyncKeychain(nil)
			if err != nil {
				c.writeLog("sync keychain error", err)
				return
			}

			c.showKeychainList()
		})

	c.loginForm.AddButton("Register",
		func() {
			err := c.app.RegisterUser(c.login, c.password)
			if err != nil {
				c.writeLog("registration error", err)
				return
			}

			k, err := keychain.NewKeychain(nil, c.app.Log)
			if err != nil {
				c.writeLog("error creation keychain", err)
				return
			}
			c.app.Keychains = append(c.app.Keychains, k)

			c.showKeychainList()
		})

	c.uiapp.SetFocus(c.loginForm)
}

func (c *UIController) writeLog(message string, err error) {
	if err != nil {
		c.log.Error(message, zap.Error(err))
		_, _ = fmt.Fprintf(c.logView, "%s: %v\n", message, err)
	} else {
		_, _ = fmt.Fprintln(c.logView, message)
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
