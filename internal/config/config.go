package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	ServerURL              string `json:"server_url"`
	Username               string `json:"username,omitempty"`
	Password               string `json:"password,omitempty"`
	AdminSecretHeaderName  string `json:"admin_secret_header_name,omitempty"`
	AdminSecretHeaderValue string `json:"admin_secret_header_value,omitempty"`
}

// Keys lists all settable config keys in display order.
var Keys = []string{
	"server",
	"username",
	"password",
	"secret-header-name",
	"secret-header-value",
}

// SensitiveKeys are masked in output.
var SensitiveKeys = map[string]bool{
	"password":            true,
	"secret-header-value": true,
}

func (c *Config) Get(key string) string {
	switch key {
	case "server":
		return c.ServerURL
	case "username":
		return c.Username
	case "password":
		return c.Password
	case "secret-header-name":
		return c.AdminSecretHeaderName
	case "secret-header-value":
		return c.AdminSecretHeaderValue
	}
	return ""
}

func (c *Config) Set(key, value string) bool {
	switch key {
	case "server":
		c.ServerURL = value
	case "username":
		c.Username = value
	case "password":
		c.Password = value
	case "secret-header-name":
		c.AdminSecretHeaderName = value
	case "secret-header-value":
		c.AdminSecretHeaderValue = value
	default:
		return false
	}
	return true
}

type Token struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(home, ".config", "kc")
	return d, os.MkdirAll(d, 0700)
}

func LoadConfig() (*Config, error) {
	d, err := dir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(d, "config.json"))
	if err != nil {
		return nil, err
	}
	var c Config
	return &c, json.Unmarshal(data, &c)
}

func SaveConfig(c *Config) error {
	d, err := dir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, "config.json"), data, 0600)
}

func LoadToken() (*Token, error) {
	d, err := dir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(d, "token.json"))
	if err != nil {
		return nil, err
	}
	var t Token
	return &t, json.Unmarshal(data, &t)
}

func SaveToken(t *Token) error {
	d, err := dir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, "token.json"), data, 0600)
}
