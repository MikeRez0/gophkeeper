/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var keychainCmd = &cobra.Command{
	Use:   "keychain",
	Short: "Local keychain control",
	Long:  `Local keychain control`,
}

func init() {
	rootCmd.AddCommand(keychainCmd)

	keychainCmd.PersistentFlags().StringVar(&keychainName, "name", "My keychain", "Keychain name")
}
