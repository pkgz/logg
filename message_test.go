package logg

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLogg_newMessage(t *testing.T) {
	tests := map[string]struct {
		level     level
		calldepth int
		flags     int
		format    format
		color     bool
	}{
		"empty": {},
		"info level": {
			level: InfoLevel,
		},
		"flags": {
			flags: LstdFlags,
		},
		"calldepth": {
			calldepth: 4,
		},
		"format": {
			format: Json,
		},
		"color": {
			color: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := newMessage(tc.level, tc.calldepth, tc.flags, tc.format, tc.color)

			if m == nil {
				t.Error("message object is nil")
				t.FailNow()
			}

			if len(m.buf) != 0 {
				t.Error("message buffer must be empty")
			}

			if m.level != tc.level {
				t.Errorf("message lvl must be equal to %d, received: %d", tc.level, m.level)
			}

			if m.calldepth != tc.calldepth {
				t.Errorf("message calldepth must be equal to %d, received: %d", tc.calldepth, m.calldepth)
			}

			if m.flags != tc.flags {
				t.Errorf("message flags must be equal to %d, received: %d", tc.flags, m.flags)
			}

			if m.format != tc.format {
				t.Errorf("message format must be equal to %d, received: %d", tc.format, m.format)
			}

			if m.color != tc.color {
				t.Errorf("message color must be equal to %v, received: %v", tc.color, m.color)
			}
		})
	}
}

func TestLogg_message_put(t *testing.T) {
	m := &message{
		buf: make([]byte, 1<<16+1),
	}

	m.put()

	mFromPool := messagePool.Get().(*message)
	if len(mFromPool.buf) >= 1<<16 {
		t.Error("message has a buf size maxSize+1, cannot be saved in sync.Pool")
	}
}

func TestLogg_message_build(t *testing.T) {
	tests := map[string]struct {
		data []byte

		level     level
		calldepth int
		flags     int
		color     bool

		pretty []byte
		json   []byte
	}{
		"empty": {
			level: EmptyLevel,
		},
		"string": {
			data:   []byte("test"),
			level:  EmptyLevel,
			pretty: []byte("test"),
			json:   []byte(`{"message": "test"}`),
		},
		"time": {
			data:   []byte("test"),
			level:  EmptyLevel,
			flags:  LstdFlags,
			pretty: []byte(fmt.Sprintf("%s test", time.Now().Format("2006-01-02 15:04:05"))),
			json:   []byte(fmt.Sprintf(`{"time": "%s", "message": "test"}`, time.Now().Format(time.RFC3339))),
		},
		"caller": {
			data:      []byte("test"),
			level:     EmptyLevel,
			flags:     Lshortfile,
			calldepth: 3,
			pretty:    []byte("$1:$2 test"),
			json:      []byte(`{"file": "$1", "line": "$2", "message": "test"}`),
		},
		"level": {
			data:   []byte("test"),
			level:  InfoLevel,
			pretty: []byte("INF test"),
			json:   []byte(`{"level": "INF", "message": "test"}`),
		},
		"color": {
			data:   []byte("test"),
			level:  ErrorLevel,
			color:  true,
			pretty: []byte(fmt.Sprintf("%s%s[1mERR%s%s test", generate(Red), escape, escapeClose, escapeClose)),
			json:   []byte(`{"level": "ERR", "message": "test"}`),
		},
		"time + level + message": {
			data:   []byte("test"),
			level:  WarningLevel,
			flags:  LstdFlags,
			pretty: []byte(fmt.Sprintf("%s WRN test", time.Now().Format("2006-01-02 15:04:05"))),
			json:   []byte(fmt.Sprintf(`{"time": "%s", "level": "WRN", "message": "test"}`, time.Now().Format(time.RFC3339))),
		},
		"time + caller + level + message": {
			data:      []byte("test"),
			level:     WarningLevel,
			flags:     LstdFlags | Lshortfile,
			calldepth: 3,
			pretty:    []byte(fmt.Sprintf("%s $1:$2 WRN test", time.Now().Format("2006-01-02 15:04:05"))),
			json:      []byte(fmt.Sprintf(`{"time": "%s", "file": "$1", "line": "$2", "level": "WRN", "message": "test"}`, time.Now().Format(time.RFC3339))),
		},
		"color + time + level + message": {
			data:  []byte("test"),
			level: WarningLevel,
			color: true,
			flags: LstdFlags,
			pretty: []byte(fmt.Sprintf("%s %s%s[1mWRN%s%s test",
				fmt.Sprintf("%s%v%s", timeColor, time.Now().Format("2006-01-02 15:04:05"), escapeClose),
				generate(HiGreen),
				escape,
				escapeClose,
				escapeClose,
			)),
			json: []byte(fmt.Sprintf(`{"time": "%s", "level": "WRN", "message": "test"}`, time.Now().Format(time.RFC3339))),
		},
		"color + caller + time + level + message": {
			data:      []byte("test"),
			level:     WarningLevel,
			color:     true,
			flags:     LstdFlags | Lshortfile,
			calldepth: 3,
			pretty: []byte(fmt.Sprintf("%s $1:$2 %s%s[1mWRN%s%s test",
				fmt.Sprintf("%s%v%s", timeColor, time.Now().Format("2006-01-02 15:04:05"), escapeClose),
				generate(HiGreen),
				escape,
				escapeClose,
				escapeClose,
			)),
			json: []byte(fmt.Sprintf(`{"time": "%s", "file": "$1", "line": "$2", "level": "WRN", "message": "test"}`, time.Now().Format(time.RFC3339))),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := newMessage(tc.level, tc.calldepth, tc.flags, Pretty, tc.color)
			m.build(tc.data)

			if tc.flags&(Lshortfile|Llongfile) != 0 {
				file, line := caller(1, tc.flags&Lshortfile != 0)
				tc.pretty = []byte(strings.Replace(string(tc.pretty), "$1", file, 1))
				tc.pretty = []byte(strings.Replace(string(tc.pretty), "$2", fmt.Sprint(line-3), 1))
			}

			if !bytes.Equal(m.buf, tc.pretty) {
				t.Errorf("wrong pretty output. Expected: %s, received: %s", string(tc.pretty), string(m.buf))
			}

			m = newMessage(tc.level, tc.calldepth, tc.flags, Json, tc.color)
			m.build(tc.data)

			if tc.flags&(Lshortfile|Llongfile) != 0 {
				file, line := caller(1, tc.flags&Lshortfile != 0)
				tc.json = []byte(strings.Replace(string(tc.json), "$1", file, 1))
				tc.json = []byte(strings.Replace(string(tc.json), "$2", fmt.Sprint(line-3), 1))
			}

			if !bytes.Equal(m.buf, tc.json) {
				t.Errorf("wrong json output. Expected: %s, received: %s", string(tc.json), string(m.buf))
			}
		})
	}
}
