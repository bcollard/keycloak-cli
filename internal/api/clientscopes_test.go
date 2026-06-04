package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreateClientScopeFull_IncludeFlag(t *testing.T) {
	var gotBody map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Location", "/admin/realms/test/client-scopes/scope-id")
		w.WriteHeader(http.StatusCreated)
	})
	c, _ := newTestClient(t, handler)

	id, err := c.CreateClientScopeFull("test", "my-scope", "desc", "openid-connect", false)
	if err != nil {
		t.Fatalf("CreateClientScopeFull: %v", err)
	}
	if id != "scope-id" {
		t.Errorf("got id=%s, want scope-id", id)
	}

	attrs, _ := gotBody["attributes"].(map[string]any)
	if attrs["include.in.token.scope"] != "false" {
		t.Errorf("include.in.token.scope should be \"false\", got %v", attrs["include.in.token.scope"])
	}
	if gotBody["protocol"] != "openid-connect" {
		t.Errorf("protocol should be openid-connect, got %v", gotBody["protocol"])
	}
}

func TestCreateClientScopeFull_DefaultsProtocol(t *testing.T) {
	var gotBody map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Location", "/admin/realms/test/client-scopes/scope-id")
		w.WriteHeader(http.StatusCreated)
	})
	c, _ := newTestClient(t, handler)

	if _, err := c.CreateClientScopeFull("test", "s", "", "", true); err != nil {
		t.Fatalf("CreateClientScopeFull: %v", err)
	}
	if gotBody["protocol"] != "openid-connect" {
		t.Errorf("empty protocol should default to openid-connect, got %v", gotBody["protocol"])
	}
}

func TestFindClientScopeByName(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]ClientScope{
			{ID: "id-a", Name: "scope-a", Protocol: "openid-connect"},
			{ID: "id-b", Name: "scope-b", Protocol: "openid-connect"},
		})
	})
	c, _ := newTestClient(t, handler)

	s, err := c.FindClientScopeByName("test", "scope-b")
	if err != nil {
		t.Fatalf("FindClientScopeByName: %v", err)
	}
	if s.ID != "id-b" {
		t.Errorf("got id=%s, want id-b", s.ID)
	}
}

func TestAddProtocolMapperToScope(t *testing.T) {
	var (
		gotPath   string
		gotMethod string
		gotBody   map[string]any
	)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Location",
			"/admin/realms/test/client-scopes/scope-id/protocol-mappers/models/new-mapper-id")
		w.WriteHeader(http.StatusCreated)
	})
	c, _ := newTestClient(t, handler)

	id, err := c.AddProtocolMapperToScope(
		"test", "scope-id",
		"aud-client-b",
		"", // protocol — should default
		"oidc-audience-mapper",
		map[string]string{
			"included.client.audience": "client-b",
			"access.token.claim":       "true",
			"id.token.claim":           "false",
		},
	)
	if err != nil {
		t.Fatalf("AddProtocolMapperToScope: %v", err)
	}
	if id != "new-mapper-id" {
		t.Errorf("got id=%s, want new-mapper-id", id)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("got method=%s, want POST", gotMethod)
	}
	wantPath := "/admin/realms/test/client-scopes/scope-id/protocol-mappers/models"
	if gotPath != wantPath {
		t.Errorf("got path=%s, want %s", gotPath, wantPath)
	}
	if gotBody["name"] != "aud-client-b" {
		t.Errorf("name = %v, want aud-client-b", gotBody["name"])
	}
	if gotBody["protocolMapper"] != "oidc-audience-mapper" {
		t.Errorf("protocolMapper = %v, want oidc-audience-mapper", gotBody["protocolMapper"])
	}
	if gotBody["protocol"] != "openid-connect" {
		t.Errorf("protocol should default to openid-connect, got %v", gotBody["protocol"])
	}
	cfg, _ := gotBody["config"].(map[string]any)
	if cfg["included.client.audience"] != "client-b" {
		t.Errorf("included.client.audience = %v, want client-b", cfg["included.client.audience"])
	}
}

func TestAddScopeToClient_PathShape(t *testing.T) {
	var gotPath, gotMethod string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	c, _ := newTestClient(t, handler)

	if err := c.AddScopeToClient("test", "client-uuid", "scope-uuid", "optional"); err != nil {
		t.Fatalf("AddScopeToClient: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("got method=%s, want PUT", gotMethod)
	}
	want := "/admin/realms/test/clients/client-uuid/optional-client-scopes/scope-uuid"
	if !strings.HasSuffix(gotPath, want) {
		t.Errorf("got path=%s, want suffix %s", gotPath, want)
	}
}
