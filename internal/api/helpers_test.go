package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bcollard/keycloak-cli/internal/config"
)

// newTestClient returns an *api.Client pointed at an httptest server with a
// long-lived token so refreshIfNeeded is a no-op.
func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := &config.Config{ServerURL: srv.URL}
	tok := &config.Token{
		AccessToken:      "test-access-token",
		RefreshToken:     "test-refresh-token",
		ExpiresAt:        time.Now().Add(time.Hour),
		RefreshExpiresAt: time.Now().Add(time.Hour),
	}
	return New(cfg, tok), srv
}
