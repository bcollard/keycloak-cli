package api

import "fmt"

type User struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email,omitempty"`
	FirstName string   `json:"firstName,omitempty"`
	LastName  string   `json:"lastName,omitempty"`
	Enabled   bool     `json:"enabled"`
	Groups    []string `json:"-"`
}

type Credential struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

type CreateUserRequest struct {
	Username    string       `json:"username"`
	Email       string       `json:"email,omitempty"`
	FirstName   string       `json:"firstName,omitempty"`
	LastName    string       `json:"lastName,omitempty"`
	Enabled     bool         `json:"enabled"`
	Credentials []Credential `json:"credentials,omitempty"`
}

func (c *Client) GetUsers(realm string) ([]User, error) {
	var users []User
	if err := c.Get(fmt.Sprintf("/admin/realms/%s/users", realm), &users); err != nil {
		return nil, err
	}
	// Enrich with group names
	for i, u := range users {
		var groups []struct {
			Name string `json:"name"`
		}
		if err := c.Get(fmt.Sprintf("/admin/realms/%s/users/%s/groups", realm, u.ID), &groups); err != nil {
			continue
		}
		for _, g := range groups {
			users[i].Groups = append(users[i].Groups, g.Name)
		}
	}
	return users, nil
}

func (c *Client) CreateUser(realm string, req CreateUserRequest) (string, error) {
	return c.PostCreated(fmt.Sprintf("/admin/realms/%s/users", realm), req)
}

func (c *Client) SetUserPassword(realm, userID, password string, temporary bool) error {
	return c.Put(fmt.Sprintf("/admin/realms/%s/users/%s/reset-password", realm, userID), Credential{
		Type:      "password",
		Value:     password,
		Temporary: temporary,
	})
}

func (c *Client) AddUserToGroup(realm, userID, groupID string) error {
	return c.Put(fmt.Sprintf("/admin/realms/%s/users/%s/groups/%s", realm, userID, groupID), nil)
}

func (c *Client) DeleteUser(realm, userID string) error {
	return c.Delete(fmt.Sprintf("/admin/realms/%s/users/%s", realm, userID))
}
