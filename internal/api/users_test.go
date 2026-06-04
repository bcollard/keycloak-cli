package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreateUserRequest_EmailVerifiedJSON(t *testing.T) {
	data, _ := json.Marshal(CreateUserRequest{
		Username:      "alex",
		Email:         "alex@example.com",
		Enabled:       true,
		EmailVerified: true,
	})
	got := string(data)
	for _, want := range []string{
		`"username":"alex"`,
		`"email":"alex@example.com"`,
		`"emailVerified":true`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("JSON missing %q\n got: %s", want, got)
		}
	}
}

func TestCreateUserRequest_OmitsEmailVerifiedWhenFalse(t *testing.T) {
	data, _ := json.Marshal(CreateUserRequest{Username: "alex"})
	if strings.Contains(string(data), `emailVerified`) {
		t.Errorf("emailVerified=false should be omitted, got: %s", data)
	}
}

func TestFindUserByUsername(t *testing.T) {
	var gotQuery string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]User{{ID: "u1", Username: "alex"}})
	})
	c, _ := newTestClient(t, handler)

	u, err := c.FindUserByUsername("test", "alex")
	if err != nil {
		t.Fatalf("FindUserByUsername: %v", err)
	}
	if u.ID != "u1" {
		t.Errorf("got id=%s, want u1", u.ID)
	}
	// Must use exact=true so a substring match (e.g. "alex" matching "alexa")
	// can't return the wrong user.
	if !strings.Contains(gotQuery, "username=alex") {
		t.Errorf("missing username param, got query: %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "exact=true") {
		t.Errorf("missing exact=true param, got query: %s", gotQuery)
	}
}

func TestFindUserByUsername_EscapesSpecialChars(t *testing.T) {
	var gotQuery string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]User{{ID: "u1", Username: "a@b.com"}})
	})
	c, _ := newTestClient(t, handler)

	_, err := c.FindUserByUsername("test", "a@b.com")
	if err != nil {
		t.Fatalf("FindUserByUsername: %v", err)
	}
	if !strings.Contains(gotQuery, "username=a%40b.com") {
		t.Errorf("@ should be url-escaped, got query: %s", gotQuery)
	}
}

func TestFindUserByUsername_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	})
	c, _ := newTestClient(t, handler)

	_, err := c.FindUserByUsername("test", "nobody")
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}
