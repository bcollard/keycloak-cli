package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the stored token",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		tokenFile := filepath.Join(home, ".config", "kc", "token.json")
		if err := os.Remove(tokenFile); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Not logged in.")
				return nil
			}
			return err
		}
		fmt.Println("Logged out.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
