package user

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bcollard/keycloak-cli/internal/api"
	"github.com/bcollard/keycloak-cli/internal/tui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var realmFlag string

var Cmd = &cobra.Command{
	Use:   "user",
	Short: "Manage Keycloak users",
}

func init() {
	Cmd.PersistentFlags().StringVarP(&realmFlag, "realm", "r", "", "Realm name (overrides REALM_NAME env var)")
	Cmd.AddCommand(listCmd, createCmd, deleteCmd)
}

func realmFromEnvOrPrompt() (string, error) {
	if realmFlag != "" {
		return realmFlag, nil
	}
	if r := os.Getenv("REALM_NAME"); r != "" {
		return r, nil
	}
	return tui.Input("Realm name", "master", true)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List users in a realm (with group memberships)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}
		users, err := client.GetUsers(realm)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(users, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		username, err := tui.Input("Username", "john.doe", true)
		if err != nil {
			return err
		}
		email, err := tui.Input("Email (optional)", "", false)
		if err != nil {
			return err
		}
		firstName, err := tui.Input("First name (optional)", "", false)
		if err != nil {
			return err
		}
		lastName, err := tui.Input("Last name (optional)", "", false)
		if err != nil {
			return err
		}

		req := api.CreateUserRequest{
			Username:  username,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			Enabled:   true,
		}

		userID, err := client.CreateUser(realm, req)
		if err != nil {
			return err
		}
		fmt.Printf("User '%s' created with ID: %s\n", username, userID)

		// Optional password
		setPassword, err := tui.Confirm("Set a password?")
		if err != nil {
			return err
		}
		if setPassword {
			password, err := tui.Password("Password")
			if err != nil {
				return err
			}
			temporary, err := tui.Confirm("Temporary password (user must change on first login)?")
			if err != nil {
				return err
			}
			if err := client.SetUserPassword(realm, userID, password, temporary); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to set password: %v\n", err)
			} else {
				fmt.Println("Password set.")
			}
		}

		// Optional group assignment
		groups, err := client.GetGroups(realm)
		if err == nil && len(groups) > 0 {
			var options []huh.Option[string]
			for _, g := range groups {
				options = append(options, huh.NewOption(g.Name, g.ID))
			}
			selectedGroups, err := tui.MultiSelect("Assign to groups (optional, press enter to skip):", options)
			if err == nil {
				for _, gID := range selectedGroups {
					if err := client.AddUserToGroup(realm, userID, gID); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to add user to group %s: %v\n", gID, err)
					} else {
						fmt.Printf("Added to group: %s\n", gID)
					}
				}
			}
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more users (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		users, err := client.GetUsers(realm)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			fmt.Println("No users found.")
			return nil
		}

		var options []huh.Option[string]
		for _, u := range users {
			label := u.Username
			parts := []string{}
			if u.Email != "" {
				parts = append(parts, u.Email)
			}
			if !u.Enabled {
				parts = append(parts, "disabled")
			}
			if len(parts) > 0 {
				label += "  (" + strings.Join(parts, ", ") + ")"
			}
			options = append(options, huh.NewOption(label, u.ID))
		}

		selected, err := tui.MultiSelect("Select users to delete:", options)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			fmt.Println("No users selected.")
			return nil
		}

		ok, err := tui.Confirm(fmt.Sprintf("Permanently delete %d user(s)?", len(selected)))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Aborted.")
			return nil
		}

		for _, id := range selected {
			if err := client.DeleteUser(realm, id); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete user %s: %v\n", id, err)
				continue
			}
			fmt.Printf("Deleted user: %s\n", id)
		}
		return nil
	},
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
