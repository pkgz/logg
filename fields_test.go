package logg

import (
	"bytes"
	stdjson "encoding/json"
	"strings"
	"testing"
)

func TestEntry_Pretty(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf)
	l.SetFlags(0)
	l.ToggleColor(false)

	l.With(F("user", "bob"), F("count", 42)).Info("logged in")

	got := strings.TrimRight(buf.String(), "\n")
	want := "INF logged in user=bob count=42"
	if got != want {
		t.Errorf("pretty entry: want %q, got %q", want, got)
	}
}

func TestEntry_PrettyQuoting(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf)
	l.SetFlags(0)
	l.ToggleColor(false)

	l.With(F("msg", "has spaces"), F("empty", "")).Warn("x")

	got := strings.TrimRight(buf.String(), "\n")
	want := `WRN x msg="has spaces" empty=""`
	if got != want {
		t.Errorf("pretty quoting: want %q, got %q", want, got)
	}
}

func TestEntry_JSON(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf)
	l.SetFlags(0)
	l.ToggleColor(false)
	l.SetFormat(Json)

	l.With(F("user", `b"b`), F("count", 42), F("ok", true)).Error("failed")

	out := bytes.TrimRight(buf.Bytes(), "\n")

	var decoded map[string]interface{}
	if err := stdjson.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("entry JSON is not valid: %v\noutput: %s", err, out)
	}
	if decoded["level"] != "ERR" {
		t.Errorf("level: got %v", decoded["level"])
	}
	if decoded["message"] != "failed" {
		t.Errorf("message: got %v", decoded["message"])
	}
	if decoded["user"] != `b"b` {
		t.Errorf("user: got %v", decoded["user"])
	}
	if decoded["count"] != float64(42) {
		t.Errorf("count: got %v (%T)", decoded["count"], decoded["count"])
	}
	if decoded["ok"] != true {
		t.Errorf("ok: got %v", decoded["ok"])
	}
}

func TestEntry_WithChaining(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf)
	l.SetFlags(0)
	l.ToggleColor(false)

	base := l.With(F("a", 1))
	base.With(F("b", 2)).Info("x")

	// base must be unchanged by the chained call.
	buf.Reset()
	base.Info("y")
	got := strings.TrimRight(buf.String(), "\n")
	if got != "INF y a=1" {
		t.Errorf("chaining mutated base entry: got %q", got)
	}
}

func TestEntry_EmptyKeySkipped(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf)
	l.SetFlags(0)
	l.ToggleColor(false)

	l.With(F("", "ignored"), F("kept", 1)).Info("x")
	got := strings.TrimRight(buf.String(), "\n")
	if got != "INF x kept=1" {
		t.Errorf("empty key not skipped: got %q", got)
	}
}
