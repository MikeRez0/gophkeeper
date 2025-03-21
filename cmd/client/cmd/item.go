package cmd

import (
	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

var itemCmd = &cobra.Command{
	Use:   "item",
	Short: "Work with keychain items",
	Long: `List, show, store itmes in keychain store.
The keychain  is selecting by name in flag [--keychain].
If there are more than one of keychains with name you should select one manualy.

Put search values to flags as query params for search item(-s) in store `,
}

var (
	filename string
)

func init() {
	rootCmd.AddCommand(itemCmd)

	itemCmd.PersistentFlags().String("label", "", "Item label (search value)")
	itemCmd.PersistentFlags().String("comment", "", "Item comment (search value)")

	itemCmd.PersistentFlags().StringVar(&filename, "file", "", "File name for binary data (for read or save)")
}

func getQueryFlags(cmd *cobra.Command) map[string]string {
	result := make(map[string]string, 5)
	addFlag := func(name string) {
		if v := cmd.Flag(name).Value.String(); v != "" {
			result[name] = v
		}
	}

	addFlag("label")
	addFlag("comment")
	addFlag("keychain")

	return result
}

func putFlagValues(e *client.CommandExecutor) *client.CommandExecutor {
	e.OfflineMode = offline
	e.KeychainPass = keychainPass
	e.Login = login
	e.Password = password
	e.Filename = filename
	return e
}
