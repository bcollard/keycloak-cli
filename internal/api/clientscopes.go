package api

import (
	"encoding/json"
	"fmt"
)

type ClientScope struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Protocol    string `json:"protocol"`
	Type        string `json:"type,omitempty"`
}

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (c *Client) GetClientScopes(realm string) ([]ClientScope, error) {
	var scopes []ClientScope
	return scopes, c.Get(fmt.Sprintf("/admin/realms/%s/client-scopes", realm), &scopes)
}

func (c *Client) CreateClientScope(realm, name, description string) (string, error) {
	return c.PostCreated(fmt.Sprintf("/admin/realms/%s/client-scopes", realm), map[string]any{
		"name":        name,
		"description": description,
		"protocol":    "openid-connect",
		"type":        "default",
		"attributes": map[string]string{
			"include.in.token.scope": "true",
		},
	})
}

func (c *Client) AddScopeToClient(realm, clientID, scopeID, scopeType string) error {
	// scopeType is "default" or "optional"
	return c.Put(fmt.Sprintf("/admin/realms/%s/clients/%s/%s-client-scopes/%s", realm, clientID, scopeType, scopeID), nil)
}
