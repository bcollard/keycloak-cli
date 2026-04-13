package config

import (
	"fmt"
	"strings"

	"github.com/bcollard/keycloak-cli/internal/config"
	"github.com/bcollard/keycloak-cli/internal/tui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage kc configuration",
}

func init() {
	Cmd.AddCommand(showCmd, listCmd, setCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("no config found — run: kc login")
		}
		for _, key := range config.Keys {
			val := cfg.Get(key)
			if val == "" {
				val = "(not set)"
			} else if config.SensitiveKeys[key] {
				val = strings.Repeat("*", 8)
			}
			fmt.Printf("%-24s %s\n", key+":", val)
		}
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available configuration keys",
	Run: func(cmd *cobra.Command, args []string) {
		for _, key := range config.Keys {
			suffix := ""
			if config.SensitiveKeys[key] {
				suffix = "  (sensitive)"
			}
			fmt.Println(key + suffix)
		}
	},
}

var setCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "Set a configuration key",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		// Validate key
		valid := false
		for _, k := range config.Keys {
			if k == key {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unknown config key %q — run 'kc config list' to see available keys", key)
		}

		// Load existing config or start fresh
		cfg, err := config.LoadConfig()
		if err != nil {
			cfg = &config.Config{}
		}

		var value string
		if len(args) == 2 {
			value = args[1]
		} else if config.SensitiveKeys[key] {
			value, err = tui.Password(fmt.Sprintf("Value for '%s'", key))
			if err != nil {
				return err
			}
		} else {
			value, err = tui.Input(fmt.Sprintf("Value for '%s'", key), "", true)
			if err != nil {
				return err
			}
		}

		cfg.Set(key, value)

		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
		fmt.Printf("Set %s\n", key)
		return nil
	},
}
