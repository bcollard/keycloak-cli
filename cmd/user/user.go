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

// create flags (non-interactive when --username is set)
var (
	createUsername       string
	createEmail          string
	createFirstName      string
	createLastName       string
	createEmailVerified  bool
	createDisabled       bool
	createPassword       string
	createTempPassword   bool
	createGroups         []string
)

// set-password flags
var (
	setPwdUserID    string
	setPwdPassword  string
	setPwdTemporary bool
)

// delete flags (non-interactive when --username or --user-id is set)
var (
	deleteUsernames []string
	deleteUserIDs   []string
)

func init() {
	Cmd.PersistentFlags().StringVarP(&realmFlag, "realm", "r", "", "Realm name (overrides REALM_NAME env var)")

	createCmd.Flags().StringVar(&createUsername, "username", "", "Username (enables non-interactive mode)")
	createCmd.Flags().StringVar(&createEmail, "email", "", "Email address")
	createCmd.Flags().StringVar(&createFirstName, "first-name", "", "First name")
	createCmd.Flags().StringVar(&createLastName, "last-name", "", "Last name")
	createCmd.Flags().BoolVar(&createEmailVerified, "email-verified", false, "Mark email as verified")
	createCmd.Flags().BoolVar(&createDisabled, "disabled", false, "Create the user disabled")
	createCmd.Flags().StringVar(&createPassword, "password", "", "Initial password (omit to skip)")
	createCmd.Flags().BoolVar(&createTempPassword, "temporary-password", false, "Mark password as temporary")
	createCmd.Flags().StringArrayVar(&createGroups, "group", nil, "Group name to assign user to (repeatable)")

	setPasswordCmd.Flags().StringVar(&setPwdUserID, "user-id", "", "User UUID (alternative to positional username)")
	setPasswordCmd.Flags().StringVar(&setPwdPassword, "password", "", "New password (required)")
	setPasswordCmd.Flags().BoolVar(&setPwdTemporary, "temporary", false, "Mark password as temporary (user must change on first login)")

	deleteCmd.Flags().StringArrayVar(&deleteUsernames, "username", nil, "Username to delete (repeatable; enables non-interactive mode)")
	deleteCmd.Flags().StringArrayVar(&deleteUserIDs, "user-id", nil, "User UUID to delete (repeatable; enables non-interactive mode)")

	Cmd.AddCommand(listCmd, createCmd, deleteCmd, setPasswordCmd)
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
	Short: "Create a new user (interactive or via flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if createUsername != "" {
			return runCreateNonInteractive(client, realm)
		}
		return runCreateInteractive(client, realm)
	},
}

func runCreateNonInteractive(client *api.Client, realm string) error {
	req := api.CreateUserRequest{
		Username:      createUsername,
		Email:         createEmail,
		FirstName:     createFirstName,
		LastName:      createLastName,
		Enabled:       !createDisabled,
		EmailVerified: createEmailVerified,
	}

	userID, err := client.CreateUser(realm, req)
	if err != nil {
		return err
	}
	fmt.Printf("User '%s' created with ID: %s\n", createUsername, userID)

	if createPassword != "" {
		if err := client.SetUserPassword(realm, userID, createPassword, createTempPassword); err != nil {
			return fmt.Errorf("user created but setting password failed: %w", err)
		}
		fmt.Println("Password set.")
	}

	if len(createGroups) > 0 {
		groups, err := client.GetGroups(realm)
		if err != nil {
			return fmt.Errorf("user created but listing groups failed: %w", err)
		}
		byName := map[string]string{}
		for _, g := range groups {
			byName[g.Name] = g.ID
		}
		for _, name := range createGroups {
			gid, ok := byName[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "Warning: group not found: %s\n", name)
				continue
			}
			if err := client.AddUserToGroup(realm, userID, gid); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to add user to group %s: %v\n", name, err)
				continue
			}
			fmt.Printf("Added to group: %s\n", name)
		}
	}
	return nil
}

func runCreateInteractive(client *api.Client, realm string) error {
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
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more users (interactive or via --username / --user-id)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if len(deleteUsernames) > 0 || len(deleteUserIDs) > 0 {
			return runDeleteNonInteractive(client, realm)
		}
		return runDeleteInteractive(client, realm)
	},
}

func runDeleteNonInteractive(client *api.Client, realm string) error {
	for _, id := range deleteUserIDs {
		if err := client.DeleteUser(realm, id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete user %s: %v\n", id, err)
			continue
		}
		fmt.Printf("Deleted user: %s\n", id)
	}
	for _, username := range deleteUsernames {
		u, err := client.FindUserByUsername(realm, username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find user '%s': %v\n", username, err)
			continue
		}
		if err := client.DeleteUser(realm, u.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete user '%s': %v\n", username, err)
			continue
		}
		fmt.Printf("Deleted user: %s (%s)\n", username, u.ID)
	}
	return nil
}

func runDeleteInteractive(client *api.Client, realm string) error {
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
}

var setPasswordCmd = &cobra.Command{
	Use:   "set-password [username]",
	Short: "Set or reset a user's password",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if setPwdPassword == "" {
			return fmt.Errorf("--password is required")
		}

		var userID string
		switch {
		case setPwdUserID != "":
			userID = setPwdUserID
		case len(args) == 1:
			u, err := client.FindUserByUsername(realm, args[0])
			if err != nil {
				return err
			}
			userID = u.ID
		default:
			return fmt.Errorf("provide a username as the positional argument or --user-id")
		}

		if err := client.SetUserPassword(realm, userID, setPwdPassword, setPwdTemporary); err != nil {
			return err
		}
		fmt.Println("Password set.")
		return nil
	},
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
