package logg

import (
	"bytes"
	stdjson "encoding/json"
	"strings"
	"testing"
)

func TestLogg_getJson(t *testing.T) {
	js := newJson()
	if js == nil {
		t.Error("js object is nil")
		t.FailNow()
	}

	if len(js.buf) == 0 {
		t.Error("js buffer is empty, but must contain a opening bracket '{'")
	}
}

func TestLogg_json_close(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		js := newJson()
		js.close()
		if string(js.buf) != `{}` {
			t.Errorf("empty json must be {}, received: %s", string(js.buf))
		}
	})

	t.Run("single field", func(t *testing.T) {
		js := newJson()
		js.addString("key", "value")
		js.close()
		if string(js.buf) != `{"key": "value"}` {
			t.Errorf("received: %s", string(js.buf))
		}
	})
}

func TestLogg_json_put(t *testing.T) {
	js := &json{
		buf: make([]byte, 1<<16+1),
	}

	js.close()
	js.put()

	jsonFromPool := jsonPool.Get().(*json)
	if len(jsonFromPool.buf) >= 1<<16 {
		t.Error("json has a buf size maxSize+1, cannot be saved in sync.Pool")
	}
}

func TestLogg_appendEscaped(t *testing.T) {
	tests := map[string]struct {
		in       string
		expected string
	}{
		"plain":         {in: "hello world", expected: "hello world"},
		"double quote":  {in: `he said "hi"`, expected: `he said \"hi\"`},
		"backslash":     {in: `a\b`, expected: `a\\b`},
		"newline":       {in: "line1\nline2", expected: `line1\nline2`},
		"carriage tab":  {in: "a\r\tb", expected: `a\r\tb`},
		"control char":  {in: "a\x00b", expected: "a\\u0000b"},
		"utf8 passthru": {in: "héllo", expected: "héllo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := string(appendEscaped(nil, []byte(tc.in)))
			if got != tc.expected {
				t.Errorf("wrong escaping. Expected: %s, received: %s", tc.expected, got)
			}
		})
	}
}

func TestLogg_buildJSON_valid(t *testing.T) {
	m := newMessage(InfoLevel, 0, 0, Json, false)
	out := m.build([]byte(`message with "quotes", a \backslash and
a newline`))

	// trailing '\n' is added by build for the writer; trim it before parsing.
	out = bytes.TrimRight(out, "\n")

	var decoded map[string]string
	if err := stdjson.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("JSON output is not valid: %v\noutput: %s", err, out)
	}

	if decoded["level"] != "INF" {
		t.Errorf("wrong level field: %q", decoded["level"])
	}
	if !strings.Contains(decoded["message"], `"quotes"`) {
		t.Errorf("message field lost its content: %q", decoded["message"])
	}
}

func TestLogg_json_fields(t *testing.T) {
	js := newJson()
	js.addString("a", "1")
	js.addBytes("b", []byte("2"))
	js.addValue("n", 42)
	js.addValue("flag", true)
	js.close()

	want := `{"a": "1", "b": "2", "n": 42, "flag": true}`
	if string(js.buf) != want {
		t.Errorf("wrong json. Expected: %s, received: %s", want, string(js.buf))
	}
}

func TestLogg_appendJSONValue(t *testing.T) {
	tests := map[string]struct {
		in       interface{}
		expected string
	}{
		"nil":      {in: nil, expected: "null"},
		"string":   {in: `a "b"`, expected: `"a \"b\""`},
		"bytes":    {in: []byte("xy"), expected: `"xy"`},
		"int":      {in: -7, expected: "-7"},
		"uint":     {in: uint(7), expected: "7"},
		"float":    {in: 1.5, expected: "1.5"},
		"bool":     {in: false, expected: "false"},
		"error":    {in: errSentinel, expected: `"boom"`},
		"stringer": {in: stringerVal{}, expected: `"S"`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := string(appendJSONValue(nil, tc.in))
			if got != tc.expected {
				t.Errorf("appendJSONValue(%v): expected %s, received %s", tc.in, tc.expected, got)
			}
		})
	}
}

var errSentinel = stringError("boom")

type stringError string

func (e stringError) Error() string { return string(e) }

type stringerVal struct{}

func (stringerVal) String() string { return "S" }
