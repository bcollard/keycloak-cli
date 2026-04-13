package group

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bcollard/keycloak-cli/internal/api"
	"github.com/bcollard/keycloak-cli/internal/tui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var realmFlag string

var Cmd = &cobra.Command{
	Use:   "group",
	Short: "Manage Keycloak groups",
}

func init() {
	Cmd.PersistentFlags().StringVarP(&realmFlag, "realm", "r", "", "Realm name (overrides REALM_NAME env var)")
	Cmd.AddCommand(listCmd, createCmd, deleteCmd, addMemberCmd)
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
	Short: "List groups in a realm (with members)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}
		groups, err := client.GetGroups(realm)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(groups, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new group (optionally as a subgroup)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		name, err := tui.Input("Group name", "my-group", true)
		if err != nil {
			return err
		}

		asSubgroup, err := tui.Confirm("Create as a subgroup of an existing group?")
		if err != nil {
			return err
		}

		var groupID string
		if asSubgroup {
			groups, err := client.GetGroups(realm)
			if err != nil {
				return err
			}
			if len(groups) == 0 {
				fmt.Println("No existing groups to use as parent.")
				return nil
			}
			var options []huh.Option[string]
			for _, g := range groups {
				options = append(options, huh.NewOption(g.Name+" ("+g.Path+")", g.ID))
			}
			parentID, err := tui.Select("Select parent group:", options)
			if err != nil {
				return err
			}
			groupID, err = client.CreateSubGroup(realm, parentID, name)
			if err != nil {
				return err
			}
		} else {
			groupID, err = client.CreateGroup(realm, name)
			if err != nil {
				return err
			}
		}

		fmt.Printf("Group '%s' created with ID: %s\n", name, groupID)
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more groups (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		groups, err := client.GetGroups(realm)
		if err != nil {
			return err
		}
		if len(groups) == 0 {
			fmt.Println("No groups found.")
			return nil
		}

		var options []huh.Option[string]
		for _, g := range groups {
			options = append(options, huh.NewOption(g.Name+" ("+g.Path+")", g.ID))
		}

		selected, err := tui.MultiSelect("Select groups to delete:", options)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			fmt.Println("No groups selected.")
			return nil
		}

		ok, err := tui.Confirm(fmt.Sprintf("Permanently delete %d group(s)?", len(selected)))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Aborted.")
			return nil
		}

		for _, id := range selected {
			if err := client.DeleteGroup(realm, id); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete group %s: %v\n", id, err)
				continue
			}
			fmt.Printf("Deleted group: %s\n", id)
		}
		return nil
	},
}

var addMemberCmd = &cobra.Command{
	Use:   "add-member",
	Short: "Add users to groups (interactive multi-select)",
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

		var userOptions []huh.Option[string]
		for _, u := range users {
			label := u.Username
			if u.Email != "" {
				label += "  (" + u.Email + ")"
			}
			userOptions = append(userOptions, huh.NewOption(label, u.ID))
		}
		selectedUsers, err := tui.MultiSelect("Select users:", userOptions)
		if err != nil {
			return err
		}
		if len(selectedUsers) == 0 {
			fmt.Println("No users selected.")
			return nil
		}

		groups, err := client.GetGroups(realm)
		if err != nil {
			return err
		}
		if len(groups) == 0 {
			fmt.Println("No groups found.")
			return nil
		}

		var groupOptions []huh.Option[string]
		for _, g := range groups {
			groupOptions = append(groupOptions, huh.NewOption(g.Name+" ("+g.Path+")", g.ID))
		}
		selectedGroups, err := tui.MultiSelect("Select groups:", groupOptions)
		if err != nil {
			return err
		}
		if len(selectedGroups) == 0 {
			fmt.Println("No groups selected.")
			return nil
		}

		for _, uID := range selectedUsers {
			for _, gID := range selectedGroups {
				if err := client.AddUserToGroup(realm, uID, gID); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to add user %s to group %s: %v\n", uID, gID, err)
				} else {
					fmt.Printf("Added user %s to group %s\n", uID, gID)
				}
			}
		}
		return nil
	},
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
