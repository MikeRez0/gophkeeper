package cmd

import (
	"context"
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

var keychainAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds new keychain",
	Long:  `Adds new local keychain with name from flag --name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := client.BootstrapApp(appConfig)
		if err != nil {
			return fmt.Errorf(cAppStartErrorText, err)
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			KeychainAdd(context.Background(), keychainName)
		if err != nil {
			return fmt.Errorf("cmd keychain add error: %w", err)
		}
		return nil
	},
}

func init() {
	keychainCmd.AddCommand(keychainAddCmd)
}
