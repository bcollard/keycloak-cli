package api

import "fmt"

type Group struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	SubGroupCount int      `json:"subGroupCount,omitempty"`
	Members       []string `json:"-"`
}

func (c *Client) GetGroups(realm string) ([]Group, error) {
	var groups []Group
	if err := c.Get(fmt.Sprintf("/admin/realms/%s/groups", realm), &groups); err != nil {
		return nil, err
	}
	for i, g := range groups {
		var members []struct {
			Username string `json:"username"`
		}
		if err := c.Get(fmt.Sprintf("/admin/realms/%s/groups/%s/members", realm, g.ID), &members); err != nil {
			continue
		}
		for _, m := range members {
			groups[i].Members = append(groups[i].Members, m.Username)
		}
	}
	return groups, nil
}

func (c *Client) CreateGroup(realm, name string) (string, error) {
	return c.PostCreated(fmt.Sprintf("/admin/realms/%s/groups", realm), map[string]any{"name": name})
}

func (c *Client) CreateSubGroup(realm, parentID, name string) (string, error) {
	return c.PostCreated(fmt.Sprintf("/admin/realms/%s/groups/%s/children", realm, parentID), map[string]any{"name": name})
}

func (c *Client) DeleteGroup(realm, groupID string) error {
	return c.Delete(fmt.Sprintf("/admin/realms/%s/groups/%s", realm, groupID))
}
