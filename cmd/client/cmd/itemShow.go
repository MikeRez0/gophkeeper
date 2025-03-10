package cmd

import (
	"context"
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

var itemShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show keychain item",
	Long:  `Show keychain item`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := client.BootstrapApp(appConfig)
		if err != nil {
			return fmt.Errorf(cAppStartErrorText, err)
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			ItemShow(context.Background(), getQueryFlags(cmd), onlySecret)
		if err != nil {
			return fmt.Errorf("cmd error: %w", err)
		}
		return nil
	},
}

var (
	onlySecret bool
)

func init() {
	itemCmd.AddCommand(itemShowCmd)

	itemShowCmd.Flags().BoolVar(&onlySecret, "only_secret", false, "Output secret only")
}
