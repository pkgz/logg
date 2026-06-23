package logg

import (
	"fmt"
	"strconv"
	"sync"
)

type json struct {
	buf  []byte
	some bool // whether at least one field has been written
}

var jsonPool = sync.Pool{
	New: func() interface{} {
		return &json{
			buf: make([]byte, 0, 500),
		}
	},
}

func newJson() *json {
	js := jsonPool.Get().(*json)
	js.buf = js.buf[:0]
	js.some = false

	js.buf = append(js.buf, '{')

	return js
}

func (js *json) close() {
	js.buf = append(js.buf, '}')
}

func (js *json) put() {
	const maxSize = 1 << 16 // 64KiB
	if cap(js.buf) > maxSize {
		return
	}

	jsonPool.Put(js)
}

// addKey writes the separator (when needed) and the quoted key followed by the
// `": "` delimiter, leaving the caller to append a fully formed value.
func (js *json) addKey(key string) {
	if js.some {
		js.buf = append(js.buf, ',', ' ')
	}
	js.some = true

	js.buf = append(js.buf, '"')
	js.buf = appendEscapedString(js.buf, key)
	js.buf = append(js.buf, '"', ':', ' ')
}

// addString writes a complete `"key": "value"` entry with the value escaped.
func (js *json) addString(key, val string) {
	js.addKey(key)
	js.buf = append(js.buf, '"')
	js.buf = appendEscapedString(js.buf, val)
	js.buf = append(js.buf, '"')
}

// addBytes writes a complete `"key": "value"` entry from a byte slice value,
// avoiding a string conversion on the hot path.
func (js *json) addBytes(key string, val []byte) {
	js.addKey(key)
	js.buf = append(js.buf, '"')
	js.buf = appendEscaped(js.buf, val)
	js.buf = append(js.buf, '"')
}

// addValue writes a complete `"key": value` entry, encoding the value
// according to its Go type.
func (js *json) addValue(key string, val interface{}) {
	js.addKey(key)
	js.buf = appendJSONValue(js.buf, val)
}

const escapeHex = "0123456789abcdef"

// appendEscaped appends s to dst, escaping characters that would otherwise
// produce invalid JSON (double quotes, backslashes and control characters).
func appendEscaped(dst, s []byte) []byte {
	for _, c := range s {
		dst = appendEscapedByte(dst, c)
	}
	return dst
}

// appendEscapedString is the string-input counterpart of appendEscaped; it
// iterates by byte index so it never allocates a []byte conversion.
func appendEscapedString(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		dst = appendEscapedByte(dst, s[i])
	}
	return dst
}

func appendEscapedByte(dst []byte, c byte) []byte {
	switch c {
	case '"':
		return append(dst, '\\', '"')
	case '\\':
		return append(dst, '\\', '\\')
	case '\n':
		return append(dst, '\\', 'n')
	case '\r':
		return append(dst, '\\', 'r')
	case '\t':
		return append(dst, '\\', 't')
	default:
		if c < 0x20 {
			return append(dst, '\\', 'u', '0', '0', escapeHex[c>>4], escapeHex[c&0xf])
		}
		return append(dst, c)
	}
}

// appendJSONValue appends v to dst encoded as a JSON value. Common scalar types
// are handled directly; anything else falls back to its fmt.Stringer/error or
// fmt.Sprint representation as a quoted string.
func appendJSONValue(dst []byte, v interface{}) []byte {
	switch val := v.(type) {
	case nil:
		return append(dst, "null"...)
	case string:
		return appendQuoted(dst, val)
	case []byte:
		dst = append(dst, '"')
		dst = appendEscaped(dst, val)
		return append(dst, '"')
	case bool:
		return strconv.AppendBool(dst, val)
	case int:
		return strconv.AppendInt(dst, int64(val), 10)
	case int8:
		return strconv.AppendInt(dst, int64(val), 10)
	case int16:
		return strconv.AppendInt(dst, int64(val), 10)
	case int32:
		return strconv.AppendInt(dst, int64(val), 10)
	case int64:
		return strconv.AppendInt(dst, val, 10)
	case uint:
		return strconv.AppendUint(dst, uint64(val), 10)
	case uint8:
		return strconv.AppendUint(dst, uint64(val), 10)
	case uint16:
		return strconv.AppendUint(dst, uint64(val), 10)
	case uint32:
		return strconv.AppendUint(dst, uint64(val), 10)
	case uint64:
		return strconv.AppendUint(dst, val, 10)
	case float32:
		return strconv.AppendFloat(dst, float64(val), 'g', -1, 32)
	case float64:
		return strconv.AppendFloat(dst, val, 'g', -1, 64)
	case error:
		return appendQuoted(dst, val.Error())
	case fmt.Stringer:
		return appendQuoted(dst, val.String())
	default:
		return appendQuoted(dst, fmt.Sprint(val))
	}
}

func appendQuoted(dst []byte, s string) []byte {
	dst = append(dst, '"')
	dst = appendEscapedString(dst, s)
	return append(dst, '"')
}
