package logg

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func newSlogTestLogger(format format) (*slog.Logger, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	l := New(buf)
	l.SetFlags(0)
	l.ToggleColor(false)
	l.SetFormat(format)
	return slog.New(l.SlogHandler()), buf
}

func TestSlog_Pretty(t *testing.T) {
	log, buf := newSlogTestLogger(Pretty)

	log.Info("hi", "user", "bob", "n", 42)

	got := strings.TrimRight(buf.String(), "\n")
	want := "INF hi user=bob n=42"
	if got != want {
		t.Errorf("slog pretty: want %q, got %q", want, got)
	}
}

func TestSlog_LevelMapping(t *testing.T) {
	tests := []struct {
		fn    func(l *slog.Logger)
		level string
	}{
		{func(l *slog.Logger) { l.Info("m") }, "INF"},
		{func(l *slog.Logger) { l.Warn("m") }, "WRN"},
		{func(l *slog.Logger) { l.Error("m") }, "ERR"},
	}

	for _, tc := range tests {
		log, buf := newSlogTestLogger(Json)
		tc.fn(log)

		var decoded map[string]interface{}
		if err := stdjson.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &decoded); err != nil {
			t.Fatalf("invalid json: %v (%s)", err, buf.String())
		}
		if decoded["level"] != tc.level {
			t.Errorf("level: want %s, got %v", tc.level, decoded["level"])
		}
	}
}

func TestSlog_DebugFilteredByDefault(t *testing.T) {
	log, buf := newSlogTestLogger(Pretty)

	log.Debug("should be dropped")

	if buf.Len() != 0 {
		t.Errorf("debug must be filtered at default min level, got: %s", buf.String())
	}
}

func TestSlog_WithAttrs(t *testing.T) {
	log, buf := newSlogTestLogger(Pretty)

	log.With("svc", "api").Info("x", "id", 7)

	got := strings.TrimRight(buf.String(), "\n")
	want := "INF x svc=api id=7"
	if got != want {
		t.Errorf("slog WithAttrs: want %q, got %q", want, got)
	}
}

func TestSlog_Groups(t *testing.T) {
	log, buf := newSlogTestLogger(Json)

	log.WithGroup("req").Info("x", slog.Int("id", 7), slog.Group("usr", slog.String("name", "bob")))

	var decoded map[string]interface{}
	if err := stdjson.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &decoded); err != nil {
		t.Fatalf("invalid json: %v (%s)", err, buf.String())
	}
	if decoded["req.id"] != float64(7) {
		t.Errorf("req.id: got %v", decoded["req.id"])
	}
	if decoded["req.usr.name"] != "bob" {
		t.Errorf("req.usr.name: got %v", decoded["req.usr.name"])
	}
}

func TestSlog_Enabled(t *testing.T) {
	buf := new(bytes.Buffer)
	l := New(buf)
	l.MinLevel(InfoLevel)
	h := l.SlogHandler()

	if h.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("debug must be disabled at InfoLevel")
	}
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("info must be enabled at InfoLevel")
	}
}
