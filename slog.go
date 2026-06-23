package logg

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// SlogHandler returns a slog.Handler that emits records through this logger,
// honoring its configured format (pretty/json), time flags, color and minimum
// level. slog attributes become structured fields; groups are flattened into
// dotted key prefixes (e.g. "req.id").
//
// The logger's file flags (Lshortfile/Llongfile) are ignored for slog records
// because logg's caller lookup cannot see through slog's call frames.
func (l *Logg) SlogHandler() slog.Handler {
	return &slogHandler{l: l}
}

// NewSlog is a convenience constructor that returns a *slog.Logger backed by a
// new logg writing to w.
func NewSlog(w io.Writer) *slog.Logger {
	return slog.New(New(w).SlogHandler())
}

type slogHandler struct {
	l      *Logg
	groups []string
	attrs  []Field
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.l.enabled(fromSlogLevel(level))
}

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	fields := make([]Field, 0, len(h.attrs)+r.NumAttrs())
	fields = append(fields, h.attrs...)

	r.Attrs(func(a slog.Attr) bool {
		fields = appendAttr(fields, h.groups, a)
		return true
	})

	h.l.writeRecord(fromSlogLevel(r.Level), fields, r.Message)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	merged := make([]Field, len(h.attrs), len(h.attrs)+len(attrs))
	copy(merged, h.attrs)
	for _, a := range attrs {
		merged = appendAttr(merged, h.groups, a)
	}

	return &slogHandler{l: h.l, groups: h.groups, attrs: merged}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	groups := make([]string, len(h.groups), len(h.groups)+1)
	copy(groups, h.groups)
	groups = append(groups, name)

	return &slogHandler{l: h.l, groups: groups, attrs: h.attrs}
}

// fromSlogLevel maps a slog.Level onto a logg level.
func fromSlogLevel(level slog.Level) level {
	switch {
	case level < slog.LevelInfo:
		return DebugLevel
	case level < slog.LevelWarn:
		return InfoLevel
	case level < slog.LevelError:
		return WarningLevel
	default:
		return ErrorLevel
	}
}

// appendAttr flattens a slog.Attr (resolving values and recursing into groups)
// into dst, prefixing keys with the current group path joined by '.'.
func appendAttr(dst []Field, groups []string, a slog.Attr) []Field {
	a.Value = a.Value.Resolve()

	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		if len(attrs) == 0 {
			return dst
		}

		// A group with an empty key is inlined without adding to the prefix.
		next := groups
		if a.Key != "" {
			next = append(append(make([]string, 0, len(groups)+1), groups...), a.Key)
		}

		for _, ga := range attrs {
			dst = appendAttr(dst, next, ga)
		}
		return dst
	}

	if a.Key == "" {
		return dst
	}

	key := a.Key
	if len(groups) > 0 {
		key = strings.Join(groups, ".") + "." + a.Key
	}

	return append(dst, F(key, a.Value.Any()))
}
