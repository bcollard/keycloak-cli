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

// create flags (non-interactive when --name is set)
var (
	createName     string
	createParent   string
	createParentID string
)

// delete flags (non-interactive when --name or --group-id is set)
var (
	deleteNames    []string
	deleteGroupIDs []string
)

// add-member flags (non-interactive when any selector is set)
var (
	addMemberUsernames []string
	addMemberUserIDs   []string
	addMemberGroups    []string
	addMemberGroupIDs  []string
)

func init() {
	Cmd.PersistentFlags().StringVarP(&realmFlag, "realm", "r", "", "Realm name (overrides REALM_NAME env var)")

	createCmd.Flags().StringVar(&createName, "name", "", "Group name (enables non-interactive mode)")
	createCmd.Flags().StringVar(&createParent, "parent", "", "Parent group name or path (creates as a subgroup)")
	createCmd.Flags().StringVar(&createParentID, "parent-id", "", "Parent group UUID (creates as a subgroup)")

	deleteCmd.Flags().StringArrayVar(&deleteNames, "name", nil, "Group name or path to delete (repeatable; enables non-interactive mode)")
	deleteCmd.Flags().StringArrayVar(&deleteGroupIDs, "group-id", nil, "Group UUID to delete (repeatable; enables non-interactive mode)")

	addMemberCmd.Flags().StringArrayVar(&addMemberUsernames, "user", nil, "Username to add (repeatable; enables non-interactive mode)")
	addMemberCmd.Flags().StringArrayVar(&addMemberUserIDs, "user-id", nil, "User UUID to add (repeatable; enables non-interactive mode)")
	addMemberCmd.Flags().StringArrayVar(&addMemberGroups, "group", nil, "Group name or path to add to (repeatable; enables non-interactive mode)")
	addMemberCmd.Flags().StringArrayVar(&addMemberGroupIDs, "group-id", nil, "Group UUID to add to (repeatable; enables non-interactive mode)")

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
	Short: "Create a new group, optionally as a subgroup (interactive or via flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if createName != "" {
			return runCreateNonInteractive(client, realm)
		}
		return runCreateInteractive(client, realm)
	},
}

func runCreateNonInteractive(client *api.Client, realm string) error {
	parentID := createParentID
	if parentID == "" && createParent != "" {
		groups, err := client.GetGroups(realm)
		if err != nil {
			return err
		}
		for _, g := range groups {
			if g.Name == createParent || g.Path == createParent {
				parentID = g.ID
				break
			}
		}
		if parentID == "" {
			return fmt.Errorf("parent group not found: %q", createParent)
		}
	}

	var (
		groupID string
		err     error
	)
	if parentID != "" {
		groupID, err = client.CreateSubGroup(realm, parentID, createName)
	} else {
		groupID, err = client.CreateGroup(realm, createName)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Group '%s' created with ID: %s\n", createName, groupID)
	return nil
}

func runCreateInteractive(client *api.Client, realm string) error {
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
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more groups (interactive or via --name / --group-id)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if len(deleteNames) > 0 || len(deleteGroupIDs) > 0 {
			return runDeleteNonInteractive(client, realm)
		}
		return runDeleteInteractive(client, realm)
	},
}

func runDeleteNonInteractive(client *api.Client, realm string) error {
	for _, id := range deleteGroupIDs {
		if err := client.DeleteGroup(realm, id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete group %s: %v\n", id, err)
			continue
		}
		fmt.Printf("Deleted group: %s\n", id)
	}
	if len(deleteNames) == 0 {
		return nil
	}

	groups, err := client.GetGroups(realm)
	if err != nil {
		return err
	}
	byName := map[string]string{}
	byPath := map[string]string{}
	for _, g := range groups {
		byName[g.Name] = g.ID
		byPath[g.Path] = g.ID
	}
	for _, n := range deleteNames {
		id, ok := byName[n]
		if !ok {
			id, ok = byPath[n]
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "Group not found: %s\n", n)
			continue
		}
		if err := client.DeleteGroup(realm, id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete group '%s': %v\n", n, err)
			continue
		}
		fmt.Printf("Deleted group: %s (%s)\n", n, id)
	}
	return nil
}

func runDeleteInteractive(client *api.Client, realm string) error {
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
}

var addMemberCmd = &cobra.Command{
	Use:   "add-member",
	Short: "Add users to groups (interactive multi-select or via flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if len(addMemberUsernames)+len(addMemberUserIDs)+len(addMemberGroups)+len(addMemberGroupIDs) > 0 {
			return runAddMemberNonInteractive(client, realm)
		}
		return runAddMemberInteractive(client, realm)
	},
}

func runAddMemberNonInteractive(client *api.Client, realm string) error {
	if len(addMemberUsernames)+len(addMemberUserIDs) == 0 {
		return fmt.Errorf("at least one --user or --user-id is required")
	}
	if len(addMemberGroups)+len(addMemberGroupIDs) == 0 {
		return fmt.Errorf("at least one --group or --group-id is required")
	}

	userIDs := append([]string{}, addMemberUserIDs...)
	for _, username := range addMemberUsernames {
		u, err := client.FindUserByUsername(realm, username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find user '%s': %v\n", username, err)
			continue
		}
		userIDs = append(userIDs, u.ID)
	}

	groupIDs := append([]string{}, addMemberGroupIDs...)
	if len(addMemberGroups) > 0 {
		groups, err := client.GetGroups(realm)
		if err != nil {
			return err
		}
		byName := map[string]string{}
		byPath := map[string]string{}
		for _, g := range groups {
			byName[g.Name] = g.ID
			byPath[g.Path] = g.ID
		}
		for _, n := range addMemberGroups {
			id, ok := byName[n]
			if !ok {
				id, ok = byPath[n]
			}
			if !ok {
				fmt.Fprintf(os.Stderr, "Group not found: %s\n", n)
				continue
			}
			groupIDs = append(groupIDs, id)
		}
	}

	for _, uID := range userIDs {
		for _, gID := range groupIDs {
			if err := client.AddUserToGroup(realm, uID, gID); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to add user %s to group %s: %v\n", uID, gID, err)
			} else {
				fmt.Printf("Added user %s to group %s\n", uID, gID)
			}
		}
	}
	return nil
}

func runAddMemberInteractive(client *api.Client, realm string) error {
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
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
