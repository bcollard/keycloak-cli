package api

type Realm struct {
	ID      string `json:"id"`
	Realm   string `json:"realm"`
	Enabled bool   `json:"enabled"`
}

func (c *Client) GetRealms() ([]Realm, error) {
	var realms []Realm
	return realms, c.Get("/admin/realms", &realms)
}

func (c *Client) CreateRealm(name string) error {
	_, err := c.PostCreated("/admin/realms", map[string]any{
		"realm":   name,
		"enabled": true,
	})
	return err
}

func (c *Client) DeleteRealm(name string) error {
	return c.Delete("/admin/realms/" + name)
}
