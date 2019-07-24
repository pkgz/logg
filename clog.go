// Package clog adding colors to your go application logs.
/*
Usage

	package main

	import (
		"github.com/exelban/clog"
		"log"
	)

	func main () {
		w := clog.Install(clog.Cyan)

		log.Print("[ERROR] error text")
	}
*/
package clog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// A Writer represents an active logging object that generates lines of
// output to an io.Writer. Each logging operation makes a single call to
// the Writer's Write method.
type Writer struct {
	out    io.Writer
	colors map[string]string
	color  string

	mu sync.Mutex
}

const escape = "\x1b"
const textBase = 30
const backgroundBase = 40

// Base attributes
const (
	Reset int = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// Install creating proxy writer for output and set it for log.
func Install(v ...interface{}) *Writer {
	w := &Writer{
		out:    os.Stderr,
		colors: make(map[string]string),
		color:  generate(v...),
	}
	log.SetOutput(w)

	w.Custom("[ERROR]", Red)
	w.Custom("[INFO]", HiYellow)
	w.Custom("[WARN]", HiGreen)
	w.Custom("[DEBUG]", HiCyan)

	return w
}

// SetOutput sets the output destination for the standard logger.
func (w *Writer) SetOutput(writer io.Writer) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.out = writer
}

// Write io.Writer implementation.
func (w *Writer) Write(b []byte) (int, error) {
	c := make([]byte, len(b))
	copy(c, b)
	var color string

	ws := sync.WaitGroup{}
	ws.Add(len(w.colors))
	for p, i := range w.colors {
		go func(p string, c string) {
			if bytes.Contains(b, []byte(p)) {
				color = c
			}
			ws.Done()
		}(p, i)
	}
	ws.Wait()

	if color == "" {
		color = w.color
	}

	if color == "" {
		n, err := w.out.Write(c)
		return n, err
	}

	w.set(color)
	n, err := w.out.Write(c)
	w.unset()

	return n, err
}

// Prefix allow to set specific colors which will be set to if prefix will be find in logging text.
func (w *Writer) Prefix(prefix string, f func(clog Colors) string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.colors[prefix] = f(&colors{})
}

// Custom allow to set custom colors for prefix.
// Accept parameters in next configuration: [textColor, backgroundColor, style].
func (w *Writer) Custom(prefix string, v ...interface{}) {
	if len(v) == 0 {
		panic(fmt.Sprintf("clog: missed configuration for %s", prefix))
	}

	switch v[0].(type) {
	case int:
	default:
		panic(fmt.Sprintf("clog: wrong configuration for %s (%v)", prefix, v))
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.colors[prefix] = generate(v...)
}

// Uninstall set log output to default (os.Stderr).
func (w *Writer) Uninstall() {
	log.SetOutput(os.Stderr)
}

func (w *Writer) set(c string) {
	str := fmt.Sprintf("%s[%sm", escape, c)
	_, _ = fmt.Fprintf(w.out, str)
}

func (w *Writer) unset() {
	_, _ = fmt.Fprintf(w.out, "%s[%dm", escape, Reset)
}

func generate(v ...interface{}) string {
	var color string

	switch len(v) {
	case 1:
		text := textBase + v[0].(int)
		color = fmt.Sprintf("%d;", text)
	case 2:
		text := textBase + v[0].(int)
		background := backgroundBase + v[1].(int)
		color = fmt.Sprintf("%d;%d;", text, background)
	case 3:
		text := textBase + v[0].(int)
		background := backgroundBase + v[1].(int)
		style := v[2].(int)
		color = fmt.Sprintf("%d;%d;%d;", style, text, background)
	}

	return color
}
