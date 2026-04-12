package realm

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bcollard/keycloak-cli/internal/api"
	"github.com/bcollard/keycloak-cli/internal/tui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "realm",
	Short: "Manage Keycloak realms",
}

func init() {
	Cmd.AddCommand(listCmd, createCmd, deleteCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all realms",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realms, err := client.GetRealms()
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(realms, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new realm",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		name, err := tui.Input("Realm name", "my-realm", true)
		if err != nil {
			return err
		}
		if err := client.CreateRealm(name); err != nil {
			return err
		}
		fmt.Printf("Realm '%s' created.\n", name)
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more realms (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireClient()
		if err != nil {
			return err
		}
		realms, err := client.GetRealms()
		if err != nil {
			return err
		}

		var options []huh.Option[string]
		for _, r := range realms {
			if r.Realm == "master" {
				continue
			}
			label := r.Realm
			if !r.Enabled {
				label += " [disabled]"
			}
			options = append(options, huh.NewOption(label, r.Realm))
		}
		if len(options) == 0 {
			fmt.Println("No deletable realms found.")
			return nil
		}

		selected, err := tui.MultiSelect("Select realms to delete:", options)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			fmt.Println("No realms selected.")
			return nil
		}

		ok, err := tui.Confirm(fmt.Sprintf("Permanently delete %d realm(s)? This cannot be undone.", len(selected)))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Aborted.")
			return nil
		}

		for _, name := range selected {
			if err := client.DeleteRealm(name); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete realm '%s': %v\n", name, err)
				continue
			}
			fmt.Printf("Deleted realm: %s\n", name)
		}
		return nil
	},
}

// requireClient is injected at registration time to avoid import cycles.
var requireClient func() (*api.Client, error)

// Register adds the realm command to the parent and injects the client factory.
func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
