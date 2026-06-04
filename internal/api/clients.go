package api

import "fmt"

type KCClient struct {
	ID                      string `json:"id"`
	ClientID                string `json:"clientId"`
	Name                    string `json:"name,omitempty"`
	Description             string `json:"description,omitempty"`
	Enabled                 bool   `json:"enabled"`
	PublicClient            bool   `json:"publicClient"`
	ServiceAccountsEnabled  bool   `json:"serviceAccountsEnabled"`
}

type CreateClientRequest struct {
	ClientID                  string            `json:"clientId"`
	Name                      string            `json:"name,omitempty"`
	Enabled                   bool              `json:"enabled"`
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool              `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`
	PublicClient              bool              `json:"publicClient"`
	Secret                    string            `json:"secret,omitempty"`
	RedirectUris              []string          `json:"redirectUris,omitempty"`
	Attributes                map[string]string `json:"attributes,omitempty"`
}

func (c *Client) GetClients(realm string) ([]KCClient, error) {
	var clients []KCClient
	return clients, c.Get(fmt.Sprintf("/admin/realms/%s/clients", realm), &clients)
}

func (c *Client) CreateClient(realm string, req CreateClientRequest) (string, error) {
	return c.PostCreated(fmt.Sprintf("/admin/realms/%s/clients", realm), req)
}

func (c *Client) DeleteClient(realm, clientID string) error {
	return c.Delete(fmt.Sprintf("/admin/realms/%s/clients/%s", realm, clientID))
}

// FindClientByClientID returns the client whose clientId field matches.
// Note: clientId is the human-readable identifier (e.g. "client-a"), distinct
// from the UUID stored in id.
func (c *Client) FindClientByClientID(realm, clientID string) (*KCClient, error) {
	clients, err := c.GetClients(realm)
	if err != nil {
		return nil, err
	}
	for i := range clients {
		if clients[i].ClientID == clientID {
			return &clients[i], nil
		}
	}
	return nil, fmt.Errorf("client not found: clientId=%q", clientID)
}

func (c *Client) GetClientInitialToken(realm string) (string, error) {
	body, err := c.PostRaw(fmt.Sprintf("/admin/realms/%s/clients-initial-access", realm), map[string]any{
		"expiration": 3600,
		"count":      15,
	})
	if err != nil {
		return "", err
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := unmarshalJSON(body, &result); err != nil {
		return "", err
	}
	return result.Token, nil
}
