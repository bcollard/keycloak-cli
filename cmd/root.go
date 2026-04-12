package cmd

import (
	"fmt"
	"os"

	"github.com/bcollard/keycloak-cli/internal/api"
	"github.com/bcollard/keycloak-cli/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kc",
	Short: "Keycloak CLI — manage realms, users, groups and clients",
}

func Execute(version, commit, date string) {
	rootCmd.Version = version + " (" + commit + ", " + date + ")"
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// requireClient loads saved config+token and returns a ready API client.
// Call this at the start of any command that needs an authenticated session.
func requireClient() (*api.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("not logged in — run: kc login")
	}
	token, err := config.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("not logged in — run: kc login")
	}
	return api.New(cfg, token), nil
}
