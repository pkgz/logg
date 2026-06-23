package logg

import (
	"fmt"
	"os"
)

// Field is a single structured key/value pair attached to a log entry.
type Field struct {
	Key   string
	Value interface{}
}

// F is a shorthand constructor for a Field.
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Entry is a logger pre-loaded with a set of structured fields. It is created
// with Logg.With and exposes the same level methods as *Logg; every message it
// emits carries the accumulated fields.
type Entry struct {
	l      *Logg
	fields []Field
}

// With returns an Entry that attaches the given fields to every message.
func (l *Logg) With(fields ...Field) *Entry {
	return &Entry{l: l, fields: fields}
}

// With returns a new Entry with additional fields, leaving the receiver
// unchanged so an Entry can be safely shared and extended.
func (e *Entry) With(fields ...Field) *Entry {
	merged := make([]Field, 0, len(e.fields)+len(fields))
	merged = append(merged, e.fields...)
	merged = append(merged, fields...)
	return &Entry{l: e.l, fields: merged}
}

func (e *Entry) Print(args ...interface{}) {
	e.l.write(1, EmptyLevel, e.fields, []byte(fmt.Sprint(args...)))
}

func (e *Entry) Printf(format string, args ...interface{}) {
	e.l.write(1, EmptyLevel, e.fields, []byte(fmt.Sprintf(format, args...)))
}

func (e *Entry) Debug(args ...interface{}) {
	e.l.write(1, DebugLevel, e.fields, []byte(fmt.Sprint(args...)))
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.l.write(1, DebugLevel, e.fields, []byte(fmt.Sprintf(format, args...)))
}

func (e *Entry) Info(args ...interface{}) {
	e.l.write(1, InfoLevel, e.fields, []byte(fmt.Sprint(args...)))
}

func (e *Entry) Infof(format string, args ...interface{}) {
	e.l.write(1, InfoLevel, e.fields, []byte(fmt.Sprintf(format, args...)))
}

func (e *Entry) Warn(args ...interface{}) {
	e.l.write(1, WarningLevel, e.fields, []byte(fmt.Sprint(args...)))
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	e.l.write(1, WarningLevel, e.fields, []byte(fmt.Sprintf(format, args...)))
}

func (e *Entry) Error(args ...interface{}) {
	e.l.write(1, ErrorLevel, e.fields, []byte(fmt.Sprint(args...)))
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.l.write(1, ErrorLevel, e.fields, []byte(fmt.Sprintf(format, args...)))
}

func (e *Entry) Panic(args ...interface{}) {
	s := fmt.Sprint(args...)
	e.l.write(1, PanicLevel, e.fields, []byte(s))
	panic(s)
}

func (e *Entry) Panicf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	e.l.write(1, PanicLevel, e.fields, []byte(s))
	panic(s)
}

func (e *Entry) Fatal(args ...interface{}) {
	e.l.write(1, FatalLevel, e.fields, []byte(fmt.Sprint(args...)))
	os.Exit(1)
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.l.write(1, FatalLevel, e.fields, []byte(fmt.Sprintf(format, args...)))
	os.Exit(1)
}

// needsQuote reports whether s must be quoted to remain an unambiguous token in
// the pretty (logfmt-style) output.
func needsQuote(s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; c == ' ' || c == '"' || c == '=' || c < 0x20 {
			return true
		}
	}
	return false
}

// appendPrettyValue appends v to dst as a logfmt-style token, quoting it only
// when it contains whitespace, quotes, '=' or control characters.
func appendPrettyValue(dst []byte, v interface{}) []byte {
	var s string
	switch val := v.(type) {
	case string:
		s = val
	case []byte:
		s = string(val)
	case error:
		s = val.Error()
	case fmt.Stringer:
		s = val.String()
	default:
		s = fmt.Sprint(val)
	}

	if needsQuote(s) {
		dst = append(dst, '"')
		dst = appendEscaped(dst, []byte(s))
		return append(dst, '"')
	}
	return append(dst, s...)
}
