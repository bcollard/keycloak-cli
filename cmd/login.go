package cmd

import (
	"fmt"
	"os"

	"github.com/bcollard/keycloak-cli/internal/api"
	"github.com/bcollard/keycloak-cli/internal/config"
	"github.com/bcollard/keycloak-cli/internal/tui"
	"github.com/spf13/cobra"
)

var (
	loginServer           string
	loginUser             string
	loginSecretHeaderName string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Keycloak and save credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load stored config as lowest-priority fallback
		stored, _ := config.LoadConfig()
		if stored == nil {
			stored = &config.Config{}
		}

		serverURL := loginServer
		if serverURL == "" {
			serverURL = os.Getenv("KC_SERVER")
		}
		if serverURL == "" {
			serverURL = stored.ServerURL
		}
		if serverURL == "" {
			var err error
			serverURL, err = tui.Input("Server URL", "https://keycloak.example.com", true)
			if err != nil {
				return err
			}
		}

		username := loginUser
		if username == "" {
			username = os.Getenv("KC_ADMIN_USER")
		}
		if username == "" {
			username = stored.Username
		}
		if username == "" {
			var err error
			username, err = tui.Input("Admin username", "admin", true)
			if err != nil {
				return err
			}
		}

		password := os.Getenv("KC_ADMIN_PASSWORD")
		if password == "" {
			password = stored.Password
		}
		if password == "" {
			var err error
			password, err = tui.Password("Admin password")
			if err != nil {
				return err
			}
		}

		secretHeaderName := loginSecretHeaderName
		if secretHeaderName == "" {
			secretHeaderName = os.Getenv("KC_ADMIN_SECRET_HEADER_NAME")
		}
		if secretHeaderName == "" {
			secretHeaderName = stored.AdminSecretHeaderName
		}

		var secretHeaderValue string
		if secretHeaderName != "" {
			secretHeaderValue = os.Getenv("KC_ADMIN_SECRET_HEADER_VALUE")
			if secretHeaderValue == "" {
				secretHeaderValue = stored.AdminSecretHeaderValue
			}
			if secretHeaderValue == "" {
				var err error
				secretHeaderValue, err = tui.Password(fmt.Sprintf("Value for header '%s'", secretHeaderName))
				if err != nil {
					return err
				}
			}
		}

		if err := api.Login(serverURL, username, password, secretHeaderName, secretHeaderValue); err != nil {
			return err
		}

		fmt.Println("Logged in successfully.")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVarP(&loginServer, "server", "s", "", "Keycloak server URL (overrides KC_SERVER env var)")
	loginCmd.Flags().StringVarP(&loginUser, "user", "u", "", "Admin username (overrides KC_ADMIN_USER env var)")
	loginCmd.Flags().StringVar(&loginSecretHeaderName, "secret-header-name", "", "Extra header name forwarded on every admin request (overrides KC_ADMIN_SECRET_HEADER_NAME env var)")
	rootCmd.AddCommand(loginCmd)
}
