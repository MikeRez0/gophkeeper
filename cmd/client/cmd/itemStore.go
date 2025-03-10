package cmd

import (
	"context"
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

var itemStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Store item to local keychain",
	Long:  `Store item to local keychain`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := client.BootstrapApp(appConfig)
		if err != nil {
			return fmt.Errorf(cAppStartErrorText, err)
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			ItemStore(context.Background(), getQueryFlags(cmd))
		if err != nil {
			return fmt.Errorf("cmd item store error: %w", err)
		}
		return nil
	},
}

func init() {
	itemCmd.AddCommand(itemStoreCmd)
}
