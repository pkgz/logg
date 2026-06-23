package logg

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGlobalLevelFuncs(t *testing.T) {
	buf := new(bytes.Buffer)
	NewGlobal(buf)
	SetFlags(0)
	ToggleColor(false)
	MinLevel(DebugLevel)

	tests := []struct {
		name string
		fn   func()
		want string
	}{
		{"Debug", func() { Debug("hello") }, "DBG hello"},
		{"Info", func() { Info("hello") }, "INF hello"},
		{"Warn", func() { Warn("hello") }, "WRN hello"},
		{"Error", func() { Error("hello") }, "ERR hello"},
		{"Infof", func() { Infof("n=%d", 42) }, "INF n=42"},
		{"Errorf", func() { Errorf("n=%d", 42) }, "ERR n=42"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.fn()
			got := strings.TrimRight(buf.String(), "\n")
			if got != tc.want {
				t.Errorf("global %s(): want %q, got %q", tc.name, tc.want, got)
			}
		})
	}
}

func TestPanic(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := New(buf)
	logger.SetFlags(0)
	logger.ToggleColor(false)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Panic must panic")
		}
		if r != "boom" {
			t.Errorf("panic value: want %q, got %v", "boom", r)
		}
		if got := strings.TrimRight(buf.String(), "\n"); got != "PNC boom" {
			t.Errorf("logged output: want %q, got %q", "PNC boom", got)
		}
	}()

	logger.Panic("boom")
	t.Fatal("Panic did not panic")
}

// TestFatal re-executes the test binary in a subprocess because Fatal calls
// os.Exit(1), which would otherwise abort the whole test run.
func TestFatal(t *testing.T) {
	if os.Getenv("LOGG_TEST_FATAL") == "1" {
		logger := New(os.Stdout)
		logger.SetFlags(0)
		logger.ToggleColor(false)
		logger.Fatal("fatal boom")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatal")
	cmd.Env = append(os.Environ(), "LOGG_TEST_FATAL=1")
	out, err := cmd.CombinedOutput()

	if e, ok := err.(*exec.ExitError); !ok || e.Success() {
		t.Fatalf("Fatal must exit with a non-zero status, err=%v", err)
	}
	if got := strings.TrimRight(string(out), "\n"); got != "FTL fatal boom" {
		t.Errorf("Fatal output: want %q, got %q", "FTL fatal boom", got)
	}
}

func TestFatalLevelParsing(t *testing.T) {
	for _, in := range []string{"FTL test", "FATAL test", "[FTL] test", "[FATAL] test"} {
		data := []byte(in)
		if lvl := defineLevel(&data); lvl != FatalLevel {
			t.Errorf("defineLevel(%q): want FatalLevel, got %d", in, lvl)
		}
	}
}
