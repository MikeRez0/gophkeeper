package cmd

import (
	"context"
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

var keychainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local keychains",
	Long:  `List local keychains`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := client.BootstrapApp(appConfig)
		if err != nil {
			return fmt.Errorf(cAppStartErrorText, err)
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			KeychainList(context.Background())
		if err != nil {
			return fmt.Errorf("cmd keychain list error: %w", err)
		}
		return nil
	},
}

func init() {
	keychainCmd.AddCommand(keychainListCmd)
}
