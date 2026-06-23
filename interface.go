package logg

import (
	"fmt"
	"io"
	"os"
)

// PRINT

func (l *Logg) Print(args ...interface{}) {
	l.write(1, EmptyLevel, nil, []byte(fmt.Sprint(args...)))
}

func (l *Logg) Printf(format string, args ...interface{}) {
	l.write(1, EmptyLevel, nil, []byte(fmt.Sprintf(format, args...)))
}

func (l *Logg) Debug(args ...interface{}) {
	l.write(1, DebugLevel, nil, []byte(fmt.Sprint(args...)))
}

func (l *Logg) Debugf(format string, args ...interface{}) {
	l.write(1, DebugLevel, nil, []byte(fmt.Sprintf(format, args...)))
}

func (l *Logg) Info(args ...interface{}) {
	l.write(1, InfoLevel, nil, []byte(fmt.Sprint(args...)))
}

func (l *Logg) Infof(format string, args ...interface{}) {
	l.write(1, InfoLevel, nil, []byte(fmt.Sprintf(format, args...)))
}

func (l *Logg) Error(args ...interface{}) {
	l.write(1, ErrorLevel, nil, []byte(fmt.Sprint(args...)))
}

func (l *Logg) Errorf(format string, args ...interface{}) {
	l.write(1, ErrorLevel, nil, []byte(fmt.Sprintf(format, args...)))
}

func (l *Logg) Warn(args ...interface{}) {
	l.write(1, WarningLevel, nil, []byte(fmt.Sprint(args...)))
}

func (l *Logg) Warnf(format string, args ...interface{}) {
	l.write(1, WarningLevel, nil, []byte(fmt.Sprintf(format, args...)))
}

// Panic logs the message at the panic level and then panics with the same
// message.
func (l *Logg) Panic(args ...interface{}) {
	s := fmt.Sprint(args...)
	l.write(1, PanicLevel, nil, []byte(s))
	panic(s)
}

// Panicf logs the formatted message at the panic level and then panics with
// the same message.
func (l *Logg) Panicf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	l.write(1, PanicLevel, nil, []byte(s))
	panic(s)
}

// Fatal logs the message at the fatal level and then calls os.Exit(1).
func (l *Logg) Fatal(args ...interface{}) {
	l.write(1, FatalLevel, nil, []byte(fmt.Sprint(args...)))
	os.Exit(1)
}

// Fatalf logs the formatted message at the fatal level and then calls
// os.Exit(1).
func (l *Logg) Fatalf(format string, args ...interface{}) {
	l.write(1, FatalLevel, nil, []byte(fmt.Sprintf(format, args...)))
	os.Exit(1)
}

// SETTINGS

func (l *Logg) DebugMode() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flags = Ldate | Ltime | Lmicroseconds | Lshortfile
	l.minLevel = DebugLevel
}

func (l *Logg) SetFormat(format format) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.format = format
}

func (l *Logg) SetFlags(flags int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flags = flags
}

func (l *Logg) SetWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

func (l *Logg) ToggleColor(value bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.color = value
}

func (l *Logg) MinLevel(level level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// Global

func Print(args ...interface{}) { logg.Print(args...) }

func Printf(format string, args ...interface{}) { logg.Printf(format, args...) }

func Debug(args ...interface{}) { logg.Debug(args...) }

func Debugf(format string, args ...interface{}) { logg.Debugf(format, args...) }

func Info(args ...interface{}) { logg.Info(args...) }

func Infof(format string, args ...interface{}) { logg.Infof(format, args...) }

func Warn(args ...interface{}) { logg.Warn(args...) }

func Warnf(format string, args ...interface{}) { logg.Warnf(format, args...) }

func Error(args ...interface{}) { logg.Error(args...) }

func Errorf(format string, args ...interface{}) { logg.Errorf(format, args...) }

func Panic(args ...interface{}) { logg.Panic(args...) }

func Panicf(format string, args ...interface{}) { logg.Panicf(format, args...) }

func Fatal(args ...interface{}) { logg.Fatal(args...) }

func Fatalf(format string, args ...interface{}) { logg.Fatalf(format, args...) }

func DebugMode() { logg.DebugMode() }

func SetFormat(format format) { logg.SetFormat(format) }

func SetFlags(flags int) { logg.SetFlags(flags) }

func SetWriter(w io.Writer) { logg.SetWriter(w) }

func ToggleColor(value bool) { logg.ToggleColor(value) }

func MinLevel(level level) { logg.MinLevel(level) }
