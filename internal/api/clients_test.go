package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreateClientRequest_JSON(t *testing.T) {
	req := CreateClientRequest{
		ClientID:                  "client-a",
		Name:                      "client-a",
		Enabled:                   true,
		DirectAccessGrantsEnabled: true,
		Secret:                    "secret-a",
		Attributes: map[string]string{
			"standard.token.exchange.enabled": "true",
		},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(data)

	mustContain := []string{
		`"clientId":"client-a"`,
		`"secret":"secret-a"`,
		`"directAccessGrantsEnabled":true`,
		`"standard.token.exchange.enabled":"true"`,
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Errorf("JSON missing %q\n got: %s", want, got)
		}
	}
}

func TestCreateClientRequest_OmitsSecretWhenEmpty(t *testing.T) {
	// Public clients must not send a "secret" field; an empty value should be
	// omitted thanks to omitempty.
	data, _ := json.Marshal(CreateClientRequest{ClientID: "pub", PublicClient: true})
	if strings.Contains(string(data), `"secret"`) {
		t.Errorf("empty secret should be omitted, got: %s", data)
	}
}

func TestFindClientByClientID(t *testing.T) {
	var gotPath string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]KCClient{
			{ID: "uuid-a", ClientID: "client-a", Enabled: true},
			{ID: "uuid-b", ClientID: "client-b", Enabled: true},
		})
	})
	c, _ := newTestClient(t, handler)

	cl, err := c.FindClientByClientID("test", "client-b")
	if err != nil {
		t.Fatalf("FindClientByClientID: %v", err)
	}
	if cl.ID != "uuid-b" {
		t.Errorf("got id=%s, want uuid-b", cl.ID)
	}
	if want := "/admin/realms/test/clients"; gotPath != want {
		t.Errorf("got path=%s, want %s", gotPath, want)
	}
}

func TestFindClientByClientID_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]KCClient{
			{ID: "uuid-a", ClientID: "client-a"},
		})
	})
	c, _ := newTestClient(t, handler)

	_, err := c.FindClientByClientID("test", "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing client, got nil")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error should mention the missing client id, got: %v", err)
	}
}

func TestCreateClient_SendsBearerAndJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
			t.Errorf("missing/wrong bearer: %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("missing/wrong content-type: %q", got)
		}
		w.Header().Set("Location", "/admin/realms/test/clients/new-client-uuid")
		w.WriteHeader(http.StatusCreated)
	})
	c, _ := newTestClient(t, handler)

	id, err := c.CreateClient("test", CreateClientRequest{ClientID: "x", Secret: "s"})
	if err != nil {
		t.Fatalf("CreateClient: %v", err)
	}
	if id != "new-client-uuid" {
		t.Errorf("got id=%q, want new-client-uuid", id)
	}
}
