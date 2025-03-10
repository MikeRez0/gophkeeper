package cmd

import (
	"context"
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

var itemListCmd = &cobra.Command{
	Use:   "list",
	Short: "List keychain items",
	Long:  `List keychain items`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := client.BootstrapApp(appConfig)
		if err != nil {
			return fmt.Errorf(cAppStartErrorText, err)
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			ItemList(context.Background(), getQueryFlags(cmd))
		if err != nil {
			return fmt.Errorf("cmd item list error: %w", err)
		}
		return nil
	},
}

func init() {
	itemCmd.AddCommand(itemListCmd)
}
