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

var keychainAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new keychain",
	Long:  `Add new keychain`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := client.BootstrapApp(appConfig)
		if err != nil {
			return err
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			KeychainAdd(context.Background(), keychainName)
		if err != nil {
			return fmt.Errorf("cmd error: %w", err)
		}
		return nil
	},
}

func init() {
	keychainCmd.AddCommand(keychainAddCmd)
}
