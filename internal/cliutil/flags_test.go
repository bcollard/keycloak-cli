package cliutil

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestAliasNormalizerResolvesAliases(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var username string
	var temporary bool
	fs.StringVar(&username, "username", "", "")
	fs.BoolVar(&temporary, "temporary", false, "")
	fs.SetNormalizeFunc(AliasNormalizer(map[string]string{
		"user":               "username",
		"temporary-password": "temporary",
	}))

	if err := fs.Parse([]string{"--user", "alice", "--temporary-password"}); err != nil {
		t.Fatalf("parse with aliases failed: %v", err)
	}
	if username != "alice" {
		t.Errorf("--user did not set username: got %q", username)
	}
	if !temporary {
		t.Error("--temporary-password did not set temporary")
	}
}

func TestAliasNormalizerCanonicalNameStillWorks(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var username string
	fs.StringVar(&username, "username", "", "")
	fs.SetNormalizeFunc(AliasNormalizer(map[string]string{"user": "username"}))

	if err := fs.Parse([]string{"--username", "bob"}); err != nil {
		t.Fatalf("parse with canonical name failed: %v", err)
	}
	if username != "bob" {
		t.Errorf("--username did not set username: got %q", username)
	}
}

// A repeatable (StringArray) flag must accumulate values passed under both the
// canonical name and an alias — this is the case that breaks if you instead
// register two flags pointing at the same slice.
func TestAliasNormalizerStringArrayAccumulates(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var users []string
	fs.StringArrayVar(&users, "username", nil, "")
	fs.SetNormalizeFunc(AliasNormalizer(map[string]string{"user": "username"}))

	if err := fs.Parse([]string{"--user", "alice", "--username", "bob"}); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(users) != 2 || users[0] != "alice" || users[1] != "bob" {
		t.Errorf("expected [alice bob], got %v", users)
	}
}
