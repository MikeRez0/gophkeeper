package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/google/uuid"
)

func (c *CommandExecutor) inputString(greeting, current string, hidden bool) string {
	fmt.Printf("%s [%s]: ", greeting, current)
	reader := bufio.NewReader(os.Stdin)
	v, _ := reader.ReadString('\n')
	v = v[:len(v)-1]
	if v == "" {
		return current
	} else {
		return v
	}
}

func (c *CommandExecutor) inputNumber(greeting string, current, maxV int) int {
	var val *int
	for val == nil {
		s := ""
		if current != 0 {
			s = strconv.Itoa(current)
		}
		s = c.inputString(greeting, s, false)
		if v, err := strconv.Atoi(s); err == nil {
			if v > 0 && v <= maxV {
				val = &v
			}
		}
	}
	return *val
}

func (c *CommandExecutor) requestKeychainPass() {
	if c.KeychainPass == "" {
		c.KeychainPass = c.inputString("Your keychain password", "", true)
	}
}

func writeItemsList(list []*domain.KCItemData) error {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)

	_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		"#", "Label", "Type", "Comment", "Time")
	if err != nil {
		return fmt.Errorf("output error: %w", err)
	}

	for i, item := range list {
		_, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			i+1, item.Label,
			item.ItemType,
			item.MetaData[domain.KCMetaKeyComment],
			item.ClientTime.Format("02.01.2006 15:04:05"))
		if err != nil {
			return fmt.Errorf("output error: %w", err)
		}
	}

	err = w.Flush()
	if err != nil {
		return fmt.Errorf("output error: %w", err)
	}
	return nil
}

func writeKeychainList(list []*domain.KCData) error {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)

	_, err := fmt.Fprintf(w, "%s\t%s\t%s\n",
		"#", "Name", "ID")
	if err != nil {
		return fmt.Errorf("output error: %w", err)
	}

	for i, k := range list {
		_, err := fmt.Fprintf(w, "%d\t%s\t%s\n",
			i+1, k.Name,
			k.ID.String(),
		)
		if err != nil {
			return fmt.Errorf("output error: %w", err)
		}
	}

	err = w.Flush()
	if err != nil {
		return fmt.Errorf("output error: %w", err)
	}
	return nil
}

type infoS struct {
	k string
	v string
}

func writeTab(list []infoS) error {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)

	for _, i := range list {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", i.k, i.v); err != nil {
			return fmt.Errorf("output error: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("output error: %w", err)
	}
	return nil
}

func (c *CommandExecutor) findKeychainItem(ctx context.Context,
	keychainID domain.KeychainID, flags map[string]string) (*domain.KCItemData, error) {
	list, err := c.queryKeychainItem(ctx, keychainID, flags)
	if err != nil {
		return nil, fmt.Errorf("items query error: %w", err)
	}

	switch len(list) {
	case 0:
		return nil, domain.ErrDataNotFound
	case 1:
		return list[0], nil
	default:
		err := writeItemsList(list)
		if err != nil {
			return nil, err
		}

		i := c.inputNumber("Select item number", 0, len(list))

		return list[i-1], nil
	}
}
func (c *CommandExecutor) queryKeychainItem(ctx context.Context,
	keychainID domain.KeychainID, flags map[string]string) ([]*domain.KCItemData, error) {
	list, err := c.app.Service.KeychainGetItemsSince(ctx,
		c.app.UserID, keychainID, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("read local keychain error: %w", err)
	}

	if len(flags) > 0 {
		l := make([]*domain.KCItemData, 0)
		for _, v := range list {
			if s, ok := flags["label"]; ok && strings.Contains(v.Label, s) {
				l = append(l, v)
				continue
			}
			if s, ok := flags["comment"]; ok && strings.Contains(v.MetaData[domain.KCMetaKeyComment], s) {
				l = append(l, v)
				continue
			}
		}
		list = l
	}
	return list, nil
}

func (c *CommandExecutor) findKeychainID(ctx context.Context, flags map[string]string) (domain.KeychainID, error) {
	list, err := c.app.Service.KeychainList(ctx, c.app.UserID)
	if err != nil {
		return domain.KeychainID(uuid.Nil), fmt.Errorf("read local keychain error: %w", err)
	}

	if s, ok := flags["keychain"]; ok {
		l := make([]*domain.KCData, 0)
		for _, k := range list {
			if strings.Contains(k.Name, s) {
				l = append(l, k)
			}
		}
		list = l
	}

	switch len(list) {
	case 0:
		return domain.KeychainID(uuid.Nil), errors.New("there is not any local keychain")
	case 1:
		return list[0].ID, nil
	default:
		err = writeKeychainList(list)
		if err != nil {
			return domain.KeychainID(uuid.Nil), err
		}

		i := c.inputNumber("Select keychain number", 0, len(list))

		return list[i-1].ID, nil
	}
}
