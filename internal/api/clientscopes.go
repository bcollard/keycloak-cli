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
	return c.CreateClientScopeFull(realm, name, description, "openid-connect", true)
}

// CreateClientScopeFull lets callers control the protocol and the
// include.in.token.scope flag (recipe scripts may want SAML or a hidden scope).
func (c *Client) CreateClientScopeFull(realm, name, description, protocol string, includeInTokenScope bool) (string, error) {
	if protocol == "" {
		protocol = "openid-connect"
	}
	includeStr := "false"
	if includeInTokenScope {
		includeStr = "true"
	}
	return c.PostCreated(fmt.Sprintf("/admin/realms/%s/client-scopes", realm), map[string]any{
		"name":        name,
		"description": description,
		"protocol":    protocol,
		"type":        "default",
		"attributes": map[string]string{
			"include.in.token.scope": includeStr,
		},
	})
}

// FindClientScopeByName returns the scope whose name matches.
func (c *Client) FindClientScopeByName(realm, name string) (*ClientScope, error) {
	scopes, err := c.GetClientScopes(realm)
	if err != nil {
		return nil, err
	}
	for i := range scopes {
		if scopes[i].Name == name {
			return &scopes[i], nil
		}
	}
	return nil, fmt.Errorf("client scope not found: name=%q", name)
}

func (c *Client) AddScopeToClient(realm, clientID, scopeID, scopeType string) error {
	// scopeType is "default" or "optional"
	return c.Put(fmt.Sprintf("/admin/realms/%s/clients/%s/%s-client-scopes/%s", realm, clientID, scopeType, scopeID), nil)
}

// AddProtocolMapperToScope attaches a protocol mapper to a client scope.
// mapperType is the Keycloak protocol-mapper id (e.g. "oidc-audience-mapper").
func (c *Client) AddProtocolMapperToScope(realm, scopeID, name, protocol, mapperType string, config map[string]string) (string, error) {
	if protocol == "" {
		protocol = "openid-connect"
	}
	return c.PostCreated(
		fmt.Sprintf("/admin/realms/%s/client-scopes/%s/protocol-mappers/models", realm, scopeID),
		map[string]any{
			"name":           name,
			"protocol":       protocol,
			"protocolMapper": mapperType,
			"config":         config,
		},
	)
}
