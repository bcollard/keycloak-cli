package clientscope

import (
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
	Use:   "client-scope",
	Short: "Manage Keycloak client scopes",
}

// create flags
var (
	createName                string
	createDescription         string
	createProtocol            string
	createIncludeInTokenScope bool
)

// add flags
var (
	addClient string
	addScope  string
	addType   string
)

// add-mapper flags
var (
	mapperScope      string
	mapperName       string
	mapperType       string
	mapperProtocol   string
	mapperConfig     []string
)

func init() {
	Cmd.PersistentFlags().StringVarP(&realmFlag, "realm", "r", "", "Realm name (overrides REALM_NAME env var)")

	createCmd.Flags().StringVar(&createName, "name", "", "Scope name (enables non-interactive mode)")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Scope description")
	createCmd.Flags().StringVar(&createProtocol, "protocol", "openid-connect", "Protocol (openid-connect | saml)")
	createCmd.Flags().BoolVar(&createIncludeInTokenScope, "include-in-token-scope", true, "Set include.in.token.scope attribute")

	addCmd.Flags().StringVar(&addClient, "client", "", "Client clientId (enables non-interactive mode)")
	addCmd.Flags().StringVar(&addScope, "scope", "", "Client-scope name")
	addCmd.Flags().StringVar(&addType, "type", "default", "Scope type: default | optional")

	addMapperCmd.Flags().StringVar(&mapperScope, "scope", "", "Client-scope name or UUID (required)")
	addMapperCmd.Flags().StringVar(&mapperName, "name", "", "Protocol-mapper instance name (required)")
	addMapperCmd.Flags().StringVar(&mapperType, "mapper-type", "", "Protocol-mapper provider id, e.g. oidc-audience-mapper (required)")
	addMapperCmd.Flags().StringVar(&mapperProtocol, "protocol", "openid-connect", "Protocol the mapper applies to")
	addMapperCmd.Flags().StringArrayVar(&mapperConfig, "config", nil, "Mapper config key=value (repeatable)")

	Cmd.AddCommand(createCmd, addCmd, addMapperCmd)
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

func parseKV(input []string) (map[string]string, error) {
	if len(input) == 0 {
		return nil, nil
	}
	m := map[string]string{}
	for _, kv := range input {
		idx := strings.Index(kv, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid key=value pair: %q", kv)
		}
		m[kv[:idx]] = kv[idx+1:]
	}
	return m, nil
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom client scope (interactive or via flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		name := createName
		description := createDescription
		if name == "" {
			name, err = tui.Input("Scope name", "my-scope", true)
			if err != nil {
				return err
			}
			description, err = tui.Input("Description (optional)", "", false)
			if err != nil {
				return err
			}
		}

		id, err := c.CreateClientScopeFull(realm, name, description, createProtocol, createIncludeInTokenScope)
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

		if addClient != "" && addScope != "" {
			return runAddNonInteractive(c, realm)
		}
		return runAddInteractive(c, realm)
	},
}

func runAddNonInteractive(c *api.Client, realm string) error {
	if addType != "default" && addType != "optional" {
		return fmt.Errorf("--type must be 'default' or 'optional'")
	}
	cl, err := c.FindClientByClientID(realm, addClient)
	if err != nil {
		return err
	}
	scope, err := c.FindClientScopeByName(realm, addScope)
	if err != nil {
		return err
	}
	if err := c.AddScopeToClient(realm, cl.ID, scope.ID, addType); err != nil {
		return err
	}
	fmt.Printf("Added %s scope '%s' to client '%s'.\n", addType, addScope, addClient)
	return nil
}

func runAddInteractive(c *api.Client, realm string) error {
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

	scopeType, err := tui.Select("Scope type:", []huh.Option[string]{
		huh.NewOption("default", "default"),
		huh.NewOption("optional", "optional"),
	})
	if err != nil {
		return err
	}

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
}

var addMapperCmd = &cobra.Command{
	Use:   "add-mapper",
	Short: "Attach a protocol mapper to an existing client scope",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if mapperScope == "" || mapperName == "" || mapperType == "" {
			return fmt.Errorf("--scope, --name and --mapper-type are required")
		}

		// Accept either a UUID or a scope name. We try name lookup first; if
		// nothing matches, fall back to using the value as a literal ID.
		scopeID := mapperScope
		if scope, err := c.FindClientScopeByName(realm, mapperScope); err == nil {
			scopeID = scope.ID
		}

		config, err := parseKV(mapperConfig)
		if err != nil {
			return err
		}

		id, err := c.AddProtocolMapperToScope(realm, scopeID, mapperName, mapperProtocol, mapperType, config)
		if err != nil {
			return err
		}
		fmt.Printf("Protocol mapper '%s' (%s) added to scope '%s' with ID: %s\n",
			mapperName, mapperType, mapperScope, id)
		return nil
	},
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
