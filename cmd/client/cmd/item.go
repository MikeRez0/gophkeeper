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

func init() {
	rootCmd.AddCommand(itemCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	itemCmd.PersistentFlags().String("label", "", "Item label (search value)")
	itemCmd.PersistentFlags().String("comment", "", "Item comment (search value)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// itemCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
	return e
}
