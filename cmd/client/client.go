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
	Short: "Create a new client (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := requireClient()
		if err != nil {
			return err
		}
		realm, err := realmFromEnvOrPrompt()
		if err != nil {
			return err
		}

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

		// Device flow attribute
		if flowSet["device"] {
			if req.Attributes == nil {
				req.Attributes = map[string]string{}
			}
			req.Attributes["oauth2.device.authorization.grant.enabled"] = "true"
		}
		// Token exchange attribute
		if flowSet["token_exchange"] {
			if req.Attributes == nil {
				req.Attributes = map[string]string{}
			}
			req.Attributes["standard.flow.enabled"] = "true"
		}

		// Redirect URIs (only for standard/implicit flows)
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
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more clients (interactive)",
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
	},
}

var requireClient func() (*api.Client, error)

func Register(parent *cobra.Command, clientFactory func() (*api.Client, error)) {
	requireClient = clientFactory
	parent.AddCommand(Cmd)
}
