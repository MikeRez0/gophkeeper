// Package cmd contains cobra-commands for client app.
package cmd

import (
	"log"
	"os"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

// Comnand rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "GophKeeper client application",
	// Long: `GophKeeper client application`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return client.RunTUI(appConfig)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	appConfig    *config.ConfigClient
	offline      bool
	login        string
	password     string
	token        string
	keychainPass string
	keychainName string
)

const cAppStartErrorText = "start app error: %w"

func init() {
	appConfig = config.NewConfigClient()

	pf := rootCmd.PersistentFlags()
	{
		pf.StringVarP(&appConfig.ConfigFile, "config", "c", appConfig.ConfigFile, "config filename")
		pf.StringVarP(&appConfig.HostString, "address", "a", appConfig.HostString, "HTTP/gRPC server endpoint")
		// pf.BoolVarP(&appConfig.GRPC, "grpc", "g", appConfig.GRPC, "Enable gRPC Mode")
		pf.IntVarP(&appConfig.SyncIntervalSeconds, "sync", "s", appConfig.SyncIntervalSeconds, "Sync interval")
		pf.StringVarP(&appConfig.App.LogLevel, "log", "l", appConfig.App.LogLevel, "Log level")

		pf.BoolVar(&offline, "offline", false, "Don't start synchronisation")
		pf.StringVar(&login, "login", "", "Login for keychain remote service")
		pf.StringVar(&password, "password", "", "Password for keychain remote service")
		pf.StringVar(&token, "token", "", "Token value for keychain remote service")
		pf.StringVar(&keychainPass, "key", "", "Password for keychain")
		pf.StringVar(&keychainName, "keychain", "", "Keychain name (search value)")
	}

	if err := appConfig.LoadConfigFile(); err != nil {
		log.Fatal(err)
		return
	}
}
