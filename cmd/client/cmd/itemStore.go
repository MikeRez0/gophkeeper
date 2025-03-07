/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
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
			return fmt.Errorf("start app error: %w", err)
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			ItemStore(context.Background(), getQueryFlags(cmd))
		if err != nil {
			return fmt.Errorf("cmd error: %w", err)
		}
		return nil
	},
}

func init() {
	itemCmd.AddCommand(itemStoreCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
