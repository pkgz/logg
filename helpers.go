package logg

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"time"
	"unsafe"
)

func write(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	if err != nil {
		return err
	}

	if n != len(b) {
		return io.ErrShortWrite
	}

	return nil
}

func caller(calldepth int, shortFile bool) (file string, line int) {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}

	if shortFile {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
	}

	return
}

func appendTimestamp(t time.Time, format format, flags int, dst []byte) []byte {
	if t.IsZero() {
		return append(dst, "0000-00-00 00:00:00"...)
	}

	if flags&LUTC != 0 {
		t = t.UTC()
	}

	t = t.Add(time.Nanosecond * 500) // To round under microsecond
	year := t.Year()
	year100 := year / 100
	year1 := year % 100
	month := t.Month()
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	micro := t.Nanosecond() / 1000

	var space byte = ' '
	if format == Json {
		space = 'T'
	}

	if flags&Ldate != 0 && flags&Ltime != 0 {
		dst = append(dst, []byte{
			digits10[year100], digits01[year100],
			digits10[year1], digits01[year1],
			'-',
			digits10[month], digits01[month],
			'-',
			digits10[day], digits01[day],
			space,
			digits10[hour], digits01[hour],
			':',
			digits10[minute], digits01[minute],
			':',
			digits10[second], digits01[second],
		}...)
	} else if flags&Ldate != 0 && flags&Ltime == 0 {
		dst = append(dst, []byte{
			digits10[year100], digits01[year100],
			digits10[year1], digits01[year1],
			'-',
			digits10[month], digits01[month],
			'-',
			digits10[day], digits01[day],
		}...)
	} else if flags&Ldate == 0 && flags&Ltime != 0 {
		dst = append(dst, []byte{
			digits10[hour], digits01[hour],
			':',
			digits10[minute], digits01[minute],
			':',
			digits10[second], digits01[second],
		}...)
	}

	if flags&Lmicroseconds != 0 {
		micro10000 := micro / 10000
		micro100 := micro / 100 % 100
		micro1 := micro % 100
		dst = append(dst, []byte{
			'.',
			digits10[micro10000], digits01[micro10000],
			digits10[micro100], digits01[micro100],
			digits10[micro1], digits01[micro1],
		}...)
	}

	if format == Json {
		_, s := t.Zone()
		if s == 0 {
			dst = append(dst, 'Z')
		} else {
			dst = append(dst, []byte{
				'+',
				digits10[s/3600], digits01[s/3600],
				':',
				'0',
				'0',
			}...)
		}
	}

	return dst
}

func defineLevel(data *[]byte) (lvl level) {
	lvl = EmptyLevel

	if len(*data) == 0 {
		return
	}

	leftPadding := 0
	rightPadding := 2
	maxRightPadding := 7
	if maxRightPadding > len(*data) {
		maxRightPadding = len(*data)
	}

	if (*data)[0] == '[' {
		leftPadding = 1
		rightPadding = 3
	}

	// levels holds the short names followed by the long names, so an index
	// found in the second half maps back to its level by subtracting numLevels.
	numLevels := len(levels) / 2

	for lvl == EmptyLevel && rightPadding < maxRightPadding {
		searchPart := (*data)[leftPadding:rightPadding]
		for i := 0; i < len(levels); i++ {
			if levels[i] == *(*string)(unsafe.Pointer(&searchPart)) {
				if i >= numLevels {
					i -= numLevels
				}
				lvl = level(i)
				break
			}
		}
		rightPadding++
	}

	return
}

func removeLevel(data []byte, lvl level) []byte {
	if len(data) == 0 || lvl == EmptyLevel {
		return data
	}

	if data[0] == '[' {
		y := bytes.IndexByte(data, ']')
		if y > 0 {
			data = (data)[1:]
			data = append(data[:y-1], data[y:]...)
		}
	}

	x := bytes.IndexByte(data, ' ')
	if x > 0 {
		data = data[x:]
	}

	for data[0] == 32 { // trip space
		data = append(data[:0], data[1:]...)
	}

	return data
}

func generate(v int) []byte {
	return []byte(fmt.Sprintf("%s[%dm", escape, 30+v))
}
