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

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register new user",
	Long:  `Register new user. Create empty local keychain if it hasn't yet`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := client.BootstrapApp(appConfig)
		if err != nil {
			return err
		}

		err = putFlagValues(client.NewCommandExecutor(app)).
			Register(context.Background())
		if err != nil {
			return fmt.Errorf("cmd error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
