package logg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger := New(nil)

	if logger.out != io.Discard {
		t.Error("logger writer must be io.Discard if nil writer provided")
	}

	if logger.color != DefaultColorOutput {
		t.Errorf("logger must have default color: %v, received: %v", DefaultColorOutput, logger.color)
	}
	if logger.format != DefaultFormat {
		t.Errorf("logger must have default format: %v, received: %v", DefaultFormat, logger.format)
	}
	if logger.flags != DefaultFlags {
		t.Errorf("logger must have default flags: %v, received: %v", DefaultFlags, logger.flags)
	}
	if logger.minLevel != DefaultMinimumLevel {
		t.Errorf("logger must have default minimum log level: %v, received: %v", DefaultMinimumLevel, logger.minLevel)
	}

	buf := new(bytes.Buffer)
	logger = New(buf)

	if logger.Writer() != buf {
		t.Error("logger writer must be buffer")
	}
}

func TestNewGlobal(t *testing.T) {
	buf := new(bytes.Buffer)
	NewGlobal(buf)

	test := "test"
	log.Print(test)

	output := readFromBuffer(buf)

	expected := appendTimestamp(time.Now(), Pretty, LstdFlags, timeColor)
	expected = append(expected, escapeClose...)
	expected = append(expected, fmt.Sprintf(" %s", test)...)

	if output != string(expected) {
		t.Errorf("global writer has wrong data. Expected: %s, received: %s", string(expected), output)
	}

	if log.Flags() != 0 {
		t.Errorf("After setting global logg Log.Flags must be 0, received: %d", log.Flags())
	}

	if reflect.TypeOf(log.Writer()) != reflect.TypeOf(logg) {
		t.Errorf("After setting global logg Log.Writer must be logg.Logg, received: %d", reflect.TypeOf(log.Writer()))
	}
}

func TestLogg_Write(t *testing.T) {
	tests := map[string]struct {
		data []byte
		val  string
	}{
		"empty": {},
		"string": {
			data: []byte("test"),
			val:  "test",
		},
		"string with new line": {
			data: []byte("test\n"),
			val:  "test",
		},
		"level": {
			data: []byte("INF test"),
			val:  "INF test",
		},
		"min level": {
			data: []byte("DBG test"),
			val:  "",
		},
	}

	buf := new(bytes.Buffer)
	logger := New(buf)
	logger.flags = 0
	logger.color = false
	logger.minLevel = InfoLevel

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			n, _ := logger.Write(tc.data)

			if n != len(tc.data) {
				t.Errorf("write n (%d) is not the same as expected (%d)", n, len(tc.data))
			}

			output := readFromBuffer(buf)
			if output != tc.val {
				t.Errorf("writed wrong data. Expected %s, received: %s", tc.val, output)
			}
		})
	}
}

func TestLogg(t *testing.T) {
	tests := map[string]struct {
		input string

		format format
		flags  int
		color  bool

		output string
	}{
		"empty": {},
		"info level short": {
			input:  "INF test",
			output: "INF test",
		},
		"info level short with escape": {
			input:  " ERR test",
			output: " ERR test",
		},
		"info level short with brackets": {
			input:  "[INF] test",
			output: "INF test",
		},
		"info level long": {
			input:  "INFO test",
			output: "INF test",
		},
		"info level long with escape": {
			input:  " ERROR test",
			output: " ERROR test",
		},
		"info level long with brackets": {
			input:  "[INFO] test",
			output: "INF test",
		},
	}

	buf := new(bytes.Buffer)
	NewGlobal(buf)

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logg.format = tc.format
			logg.flags = tc.flags
			logg.color = tc.color

			log.Print(tc.input)

			output := readFromBuffer(buf)
			if output != tc.output {
				t.Errorf("print error. Expected %s, received: %s", tc.output, output)
			}

			logg.format = DefaultFormat
			logg.flags = DefaultFlags
			logg.color = DefaultColorOutput
		})
	}
}

func TestDebugMode(t *testing.T) {
	t.Run("no debug mode", func(t *testing.T) {
		buf := new(bytes.Buffer)
		logger := New(buf)
		logger.Print("[DEBUG] test")

		if buf.Len() != 0 {
			t.Errorf("Debug message must be missed. Received: %s", readFromBuffer(buf))
		}

		logger.Print("[INFO] test")

		if buf.Len() == 0 {
			t.Errorf("Info message must writed missed. Received: %s", readFromBuffer(buf))
		}
	})

	t.Run("debug mode", func(t *testing.T) {
		buf := new(bytes.Buffer)
		logger := New(buf)
		logger.DebugMode()

		logger.Print("[DEBUG] test")

		if buf.Len() == 0 {
			t.Errorf("Debug message must be missed. Received: %s", readFromBuffer(buf))
		}
	})
}

func readFromBuffer(buf *bytes.Buffer) string {
	readBuf, _ := io.ReadAll(buf)
	return strings.Replace(string(readBuf), "\n", "", 1)
}
