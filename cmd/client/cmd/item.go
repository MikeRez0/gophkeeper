/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

// itemCmd represents the item command
var itemCmd = &cobra.Command{
	Use:   "item",
	Short: "Work with keychain items",
	Long:  `Work with keychain items`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("item called")
	// },
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
