package cmd

import (
	"fmt"
	"os"

	"github.com/bcollard/keycloak-cli/internal/api"
	"github.com/bcollard/keycloak-cli/internal/tui"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Keycloak and save credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL := os.Getenv("KC_SERVER")
		if serverURL == "" {
			var err error
			serverURL, err = tui.Input("Server URL", "https://keycloak.example.com", true)
			if err != nil {
				return err
			}
		}

		username := os.Getenv("KC_ADMIN_USER")
		if username == "" {
			username = "admin"
			var err error
			username, err = tui.Input("Admin username", "admin", true)
			if err != nil {
				return err
			}
		}

		password := os.Getenv("KC_ADMIN_PASSWORD")
		if password == "" {
			var err error
			password, err = tui.Password("Admin password")
			if err != nil {
				return err
			}
		}

		secretHeader := os.Getenv("KC_ADMIN_SECRET_HEADER")

		if err := api.Login(serverURL, username, password, secretHeader); err != nil {
			return err
		}

		fmt.Println("Logged in successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
