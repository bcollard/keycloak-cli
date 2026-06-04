package clientscope

import "testing"

func TestParseKV_Valid(t *testing.T) {
	m, err := parseKV([]string{"a=b", "c=d"})
	if err != nil {
		t.Fatalf("parseKV: %v", err)
	}
	if m["a"] != "b" || m["c"] != "d" {
		t.Errorf("got %v, want {a:b c:d}", m)
	}
}

func TestParseKV_Empty(t *testing.T) {
	m, err := parseKV(nil)
	if err != nil {
		t.Fatalf("parseKV(nil): %v", err)
	}
	if m != nil {
		t.Errorf("got %v, want nil for empty input", m)
	}
}

func TestParseKV_SplitsOnFirstEquals(t *testing.T) {
	// audience values often contain '=' (think base64 etc); we must keep
	// everything after the first '=' verbatim.
	m, err := parseKV([]string{"key=a=b=c"})
	if err != nil {
		t.Fatalf("parseKV: %v", err)
	}
	if m["key"] != "a=b=c" {
		t.Errorf("got %q, want a=b=c", m["key"])
	}
}

func TestParseKV_RejectsMissingEquals(t *testing.T) {
	if _, err := parseKV([]string{"nokv"}); err == nil {
		t.Fatal("expected error for missing '='")
	}
}

func TestParseKV_RejectsLeadingEquals(t *testing.T) {
	if _, err := parseKV([]string{"=v"}); err == nil {
		t.Fatal("expected error for empty key")
	}
}
