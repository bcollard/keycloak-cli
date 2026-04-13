package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bcollard/keycloak-cli/internal/config"
)

type Client struct {
	httpClient        *http.Client
	serverURL         string
	secretHeaderName  string
	secretHeaderValue string
	token             *config.Token
}

func New(cfg *config.Config, token *config.Token) *Client {
	return &Client{
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		serverURL:         strings.TrimRight(cfg.ServerURL, "/"),
		secretHeaderName:  cfg.AdminSecretHeaderName,
		secretHeaderValue: cfg.AdminSecretHeaderValue,
		token:             token,
	}
}

// Login obtains a token via ROPC and saves both config and token.
func Login(serverURL, username, password, secretHeaderName, secretHeaderValue string) error {
	serverURL = strings.TrimRight(serverURL, "/")
	tokenURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", serverURL)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", "admin-cli")
	data.Set("username", username)
	data.Set("password", password)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed (%d): %s", resp.StatusCode, body)
	}

	var raw struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	now := time.Now()
	t := &config.Token{
		AccessToken:      raw.AccessToken,
		RefreshToken:     raw.RefreshToken,
		ExpiresAt:        now.Add(time.Duration(raw.ExpiresIn) * time.Second),
		RefreshExpiresAt: now.Add(time.Duration(raw.RefreshExpiresIn) * time.Second),
	}

	cfg := &config.Config{ServerURL: serverURL, AdminSecretHeaderName: secretHeaderName, AdminSecretHeaderValue: secretHeaderValue}
	if err := config.SaveConfig(cfg); err != nil {
		return err
	}
	return config.SaveToken(t)
}

func (c *Client) refreshIfNeeded() error {
	if time.Now().Before(c.token.ExpiresAt.Add(-30 * time.Second)) {
		return nil
	}
	if time.Now().After(c.token.RefreshExpiresAt) {
		return fmt.Errorf("session expired, please run: kc login")
	}

	tokenURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", c.serverURL)
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", "admin-cli")
	data.Set("refresh_token", c.token.RefreshToken)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, body)
	}

	var raw struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}

	now := time.Now()
	c.token.AccessToken = raw.AccessToken
	c.token.RefreshToken = raw.RefreshToken
	c.token.ExpiresAt = now.Add(time.Duration(raw.ExpiresIn) * time.Second)
	c.token.RefreshExpiresAt = now.Add(time.Duration(raw.RefreshExpiresIn) * time.Second)
	return config.SaveToken(c.token)
}

func (c *Client) do(method, path string, body any) ([]byte, int, error) {
	if err := c.refreshIfNeeded(); err != nil {
		return nil, 0, err
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.serverURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.secretHeaderName != "" && c.secretHeaderValue != "" {
		req.Header.Set(c.secretHeaderName, c.secretHeaderValue)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, nil
}

// checkResponse returns a clear error when the server returns a non-JSON body
// (e.g. an HTML error page from a reverse proxy) or a 4xx/5xx status.
func checkResponse(method, path string, status int, body []byte) error {
	if len(body) > 0 && body[0] == '<' {
		hint := ""
		if status == http.StatusOK {
			// LB returned 200 + HTML — likely a missing or invalid secret header.
			hint = " (the server returned an HTML page instead of JSON — " +
				"you may need to re-run 'kc login' with --secret-header-name and KC_ADMIN_SECRET_HEADER_VALUE)"
		}
		return fmt.Errorf("%s %s failed (%d): server returned HTML instead of JSON%s",
			method, path, status, hint)
	}
	switch {
	case status == http.StatusUnauthorized:
		return fmt.Errorf("%s %s failed (401 Unauthorized): token may be expired — run: kc login", method, path)
	case status == http.StatusForbidden:
		return fmt.Errorf("%s %s failed (403 Forbidden): check your credentials and secret header — run: kc login", method, path)
	case status >= 400:
		return fmt.Errorf("%s %s failed (%d): %s", method, path, status, body)
	}
	return nil
}

func (c *Client) Get(path string, out any) error {
	body, status, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	if err := checkResponse(http.MethodGet, path, status, body); err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func (c *Client) Post(path string, payload any) (string, error) {
	body, status, err := c.do(http.MethodPost, path, payload)
	if err != nil {
		return "", err
	}
	if err := checkResponse(http.MethodPost, path, status, body); err != nil {
		return "", err
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return "", fmt.Errorf("POST %s failed (%d): %s", path, status, body)
	}
	// Extract ID from Location header is not available here; return body for callers that need it
	var result map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &result)
		if id, ok := result["id"].(string); ok {
			return id, nil
		}
	}
	return "", nil
}

func (c *Client) PostRaw(path string, payload any) ([]byte, error) {
	body, status, err := c.do(http.MethodPost, path, payload)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(http.MethodPost, path, status, body); err != nil {
		return nil, err
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return nil, fmt.Errorf("POST %s failed (%d): %s", path, status, body)
	}
	return body, nil
}

func (c *Client) Put(path string, payload any) error {
	body, status, err := c.do(http.MethodPut, path, payload)
	if err != nil {
		return err
	}
	if err := checkResponse(http.MethodPut, path, status, body); err != nil {
		return err
	}
	if status != http.StatusNoContent && status != http.StatusOK {
		return fmt.Errorf("PUT %s failed (%d): %s", path, status, body)
	}
	return nil
}

func (c *Client) Delete(path string) error {
	body, status, err := c.do(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	if err := checkResponse(http.MethodDelete, path, status, body); err != nil {
		return err
	}
	if status != http.StatusNoContent && status != http.StatusOK {
		return fmt.Errorf("DELETE %s failed (%d): %s", path, status, body)
	}
	return nil
}

// PostCreated issues a POST and extracts the created resource ID from the Location header.
func (c *Client) PostCreated(path string, payload any) (string, error) {
	if err := c.refreshIfNeeded(); err != nil {
		return "", err
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.serverURL+path, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	if c.secretHeaderName != "" && c.secretHeaderValue != "" {
		req.Header.Set(c.secretHeaderName, c.secretHeaderValue)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("POST %s failed (%d): %s", path, resp.StatusCode, body)
	}

	// Try Location header first
	if loc := resp.Header.Get("Location"); loc != "" {
		parts := strings.Split(loc, "/")
		return parts[len(parts)-1], nil
	}
	// Fall back to body id field
	var result map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &result); err == nil {
			if id, ok := result["id"].(string); ok {
				return id, nil
			}
		}
	}
	return "", nil
}
