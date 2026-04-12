package cmd

import (
	"github.com/bcollard/keycloak-cli/cmd/client"
	"github.com/bcollard/keycloak-cli/cmd/clientscope"
	"github.com/bcollard/keycloak-cli/cmd/group"
	"github.com/bcollard/keycloak-cli/cmd/realm"
	"github.com/bcollard/keycloak-cli/cmd/user"
)

func init() {
	realm.Register(rootCmd, requireClient)
	user.Register(rootCmd, requireClient)
	group.Register(rootCmd, requireClient)
	client.Register(rootCmd, requireClient)
	clientscope.Register(rootCmd, requireClient)
}
