package client

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
	Use:   "client",
	Short: "Manage Keycloak clients",
}

// delete flags (non-interactive when --client-id is set)
var deleteClientIDs []string

// create flags (non-interactive when --client-id is set)
var (
	createClientID            string
	createName                string
	createSecret              string
	createPublic              bool
	createStandardFlow        bool
	createImplicitFlow        bool
	createDirectAccessGrants  bool
	createServiceAccounts     bool
	createDeviceFlow          bool
	createTokenExchange       bool
	createRedirectURIs        []string
	createAttributes          []string
	createDisabled            bool
)

func init() {
	Cmd.PersistentFlags().StringVarP(&realmFlag, "realm", "r", "", "Realm name (overrides REALM_NAME env var)")

	createCmd.Flags().StringVar(&createClientID, "client-id", "", "Client ID (enables non-interactive mode)")
	createCmd.Flags().StringVar(&createName, "name", "", "Client display name (defaults to client-id)")
	createCmd.Flags().StringVar(&createSecret, "secret", "", "Fixed client secret (confidential clients)")
	createCmd.Flags().BoolVar(&createPublic, "public", false, "Public client (no secret)")
	createCmd.Flags().BoolVar(&createStandardFlow, "standard-flow", false, "Enable Authorization Code flow")
	createCmd.Flags().BoolVar(&createImplicitFlow, "implicit-flow", false, "Enable Implicit flow")
	createCmd.Flags().BoolVar(&createDirectAccessGrants, "direct-access-grants", false, "Enable ROPC (direct access grants)")
	createCmd.Flags().BoolVar(&createServiceAccounts, "service-accounts", false, "Enable Client Credentials grant")
	createCmd.Flags().BoolVar(&createDeviceFlow, "device-flow", false, "Enable Device Authorization grant")
	createCmd.Flags().BoolVar(&createTokenExchange, "token-exchange", false, "Enable Standard Token Exchange (Keycloak 26+)")
	createCmd.Flags().StringArrayVar(&createRedirectURIs, "redirect-uri", nil, "Redirect URI (repeatable)")
	createCmd.Flags().StringArrayVar(&createAttributes, "attribute", nil, "Extra attribute key=value (repeatable)")
	createCmd.Flags().BoolVar(&createDisabled, "disabled", false, "Create the client disabled")

	deleteCmd.Flags().StringArrayVar(&deleteClientIDs, "client-id", nil, "Client clientId to delete (repeatable; enables non-interactive mode)")

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

// parseKV parses ["a=b","c=d"] into {"a":"b","c":"d"}.
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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List clients in a realm",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}
		clients, err := c.GetClients(realm)
		if err != nil {
			return err
		}
		out, _ := json.MarshalIndent(clients, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new client (interactive or via flags)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if createClientID != "" {
			return runCreateNonInteractive(c, realm)
		}
		return runCreateInteractive(c, realm)
	},
}

func runCreateNonInteractive(c *api.Client, realm string) error {
	name := createName
	if name == "" {
		name = createClientID
	}

	attrs, err := parseKV(createAttributes)
	if err != nil {
		return err
	}
	if createDeviceFlow {
		if attrs == nil {
			attrs = map[string]string{}
		}
		attrs["oauth2.device.authorization.grant.enabled"] = "true"
	}
	if createTokenExchange {
		if attrs == nil {
			attrs = map[string]string{}
		}
		attrs["standard.token.exchange.enabled"] = "true"
	}

	req := api.CreateClientRequest{
		ClientID:                  createClientID,
		Name:                      name,
		Enabled:                   !createDisabled,
		StandardFlowEnabled:       createStandardFlow,
		ImplicitFlowEnabled:       createImplicitFlow,
		DirectAccessGrantsEnabled: createDirectAccessGrants,
		ServiceAccountsEnabled:    createServiceAccounts,
		PublicClient:              createPublic,
		Secret:                    createSecret,
		RedirectUris:              createRedirectURIs,
		Attributes:                attrs,
	}

	id, err := c.CreateClient(realm, req)
	if err != nil {
		return err
	}
	fmt.Printf("Client '%s' created with ID: %s\n", createClientID, id)
	return nil
}

func runCreateInteractive(c *api.Client, realm string) error {
	clientID, err := tui.Input("Client ID", "my-app", true)
	if err != nil {
		return err
	}
	name, err := tui.Input("Client name (optional)", "", false)
	if err != nil {
		return err
	}
	if name == "" {
		name = clientID
	}

	// Flow selection
	flowOptions := []huh.Option[string]{
		huh.NewOption("Authorization Code Flow", "standard"),
		huh.NewOption("Implicit Flow", "implicit"),
		huh.NewOption("Resource Owner Password Credentials", "direct"),
		huh.NewOption("Client Credentials Grant", "service_accounts"),
		huh.NewOption("Device Authorization Grant", "device"),
		huh.NewOption("Token Exchange", "token_exchange"),
	}
	selectedFlows, err := tui.MultiSelect("Enable OAuth2/OIDC flows:", flowOptions)
	if err != nil {
		return err
	}

	flowSet := map[string]bool{}
	for _, f := range selectedFlows {
		flowSet[f] = true
	}

	req := api.CreateClientRequest{
		ClientID:                  clientID,
		Name:                      name,
		Enabled:                   true,
		StandardFlowEnabled:       flowSet["standard"],
		ImplicitFlowEnabled:       flowSet["implicit"],
		DirectAccessGrantsEnabled: flowSet["direct"],
		ServiceAccountsEnabled:    flowSet["service_accounts"],
		PublicClient:              !flowSet["service_accounts"],
	}

	if flowSet["device"] {
		if req.Attributes == nil {
			req.Attributes = map[string]string{}
		}
		req.Attributes["oauth2.device.authorization.grant.enabled"] = "true"
	}
	if flowSet["token_exchange"] {
		if req.Attributes == nil {
			req.Attributes = map[string]string{}
		}
		req.Attributes["standard.token.exchange.enabled"] = "true"
	}

	if flowSet["standard"] || flowSet["implicit"] {
		uriText, err := tui.Text("Redirect URIs (one per line or comma-separated)", "https://app.example.com/callback")
		if err != nil {
			return err
		}
		if uriText != "" {
			raw := strings.ReplaceAll(uriText, ",", "\n")
			for _, u := range strings.Split(raw, "\n") {
				u = strings.TrimSpace(u)
				if u != "" {
					req.RedirectUris = append(req.RedirectUris, u)
				}
			}
		}
	}

	id, err := c.CreateClient(realm, req)
	if err != nil {
		return err
	}
	fmt.Printf("Client '%s' created with ID: %s\n", clientID, id)
	return nil
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more clients (interactive or via --client-id)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

		if len(deleteClientIDs) > 0 {
			return runDeleteNonInteractive(c, realm)
		}
		return runDeleteInteractive(c, realm)
	},
}

func runDeleteNonInteractive(c *api.Client, realm string) error {
	for _, clientID := range deleteClientIDs {
		cl, err := c.FindClientByClientID(realm, clientID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find client '%s': %v\n", clientID, err)
			continue
		}
		if err := c.DeleteClient(realm, cl.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete client '%s': %v\n", clientID, err)
			continue
		}
		fmt.Printf("Deleted client: %s (%s)\n", clientID, cl.ID)
	}
	return nil
}

func runDeleteInteractive(c *api.Client, realm string) error {
	clients, err := c.GetClients(realm)
	if err != nil {
		return err
	}
	if len(clients) == 0 {
		fmt.Println("No clients found.")
		return nil
	}

	var options []huh.Option[string]
	for _, cl := range clients {
		label := cl.ClientID
		if cl.Name != "" && cl.Name != cl.ClientID {
			label += "  (" + cl.Name + ")"
		}
		if !cl.Enabled {
			label += "  [disabled]"
		}
		options = append(options, huh.NewOption(label, cl.ID))
	}

	selected, err := tui.MultiSelect("Select clients to delete:", options)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Println("No clients selected.")
		return nil
	}

	ok, err := tui.Confirm(fmt.Sprintf("Permanently delete %d client(s)? This cannot be undone.", len(selected)))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Aborted.")
		return nil
	}

	for _, id := range selected {
		if err := c.DeleteClient(realm, id); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete client %s: %v\n", id, err)
			continue
		}
		fmt.Printf("Deleted client: %s\n", id)
	}
	return nil
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
