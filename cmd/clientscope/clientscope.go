package clientscope

import (
	"fmt"
	"os"

	"github.com/bcollard/keycloak-cli/internal/api"
	"github.com/bcollard/keycloak-cli/internal/tui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "client-scope",
	Short: "Manage Keycloak client scopes",
}

func init() {
	Cmd.AddCommand(createCmd, addCmd)
}

func realmFromEnvOrPrompt() (string, error) {
	if r := os.Getenv("REALM_NAME"); r != "" {
		return r, nil
	}
	return tui.Input("Realm name", "master", true)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom client scope (type=default, included in token scope)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		name, err := tui.Input("Scope name", "my-scope", true)
		if err != nil {
			return err
		}
		description, err := tui.Input("Description (optional)", "", false)
		if err != nil {
			return err
		}

		id, err := c.CreateClientScope(realm, name, description)
		if err != nil {
			return err
		}
		fmt.Printf("Client scope '%s' created with ID: %s\n", name, id)
		return nil
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a scope (default or optional) to an existing client",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		// Select client
		clients, err := c.GetClients(realm)
		if err != nil {
			return err
		}
		if len(clients) == 0 {
			fmt.Println("No clients found.")
			return nil
		}
		var clientOptions []huh.Option[string]
		for _, cl := range clients {
			label := cl.ClientID
			if cl.Name != "" && cl.Name != cl.ClientID {
				label += "  (" + cl.Name + ")"
			}
			clientOptions = append(clientOptions, huh.NewOption(label, cl.ID))
		}
		clientID, err := tui.Select("Select a client:", clientOptions)
		if err != nil {
			return err
		}

		// Select scope type
		scopeType, err := tui.Select("Scope type:", []huh.Option[string]{
			huh.NewOption("default", "default"),
			huh.NewOption("optional", "optional"),
		})
		if err != nil {
			return err
		}

		// Select scope
		scopes, err := c.GetClientScopes(realm)
		if err != nil {
			return err
		}
		if len(scopes) == 0 {
			fmt.Println("No client scopes found.")
			return nil
		}
		var scopeOptions []huh.Option[string]
		for _, s := range scopes {
			label := s.Name
			if s.Description != "" {
				label += "  (" + s.Description + ")"
			}
			label += "  [" + s.Protocol + "]"
			scopeOptions = append(scopeOptions, huh.NewOption(label, s.ID))
		}
		scopeID, err := tui.Select(fmt.Sprintf("Select scope to add as %s:", scopeType), scopeOptions)
		if err != nil {
			return err
		}

		if err := c.AddScopeToClient(realm, clientID, scopeID, scopeType); err != nil {
			return err
		}

		// Resolve names for output
		clientName := clientID
		for _, cl := range clients {
			if cl.ID == clientID {
				clientName = cl.ClientID
				break
			}
		}
		scopeName := scopeID
		for _, s := range scopes {
			if s.ID == scopeID {
				scopeName = s.Name
				break
			}
		}
		fmt.Printf("Added %s scope '%s' to client '%s'.\n", scopeType, scopeName, clientName)
		return nil
	},
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
