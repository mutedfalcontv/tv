// Package output centralizes machine-readable CLI output.
//
// When JSON mode is enabled every command emits exactly one JSON value on
// stdout. Errors are emitted as a JSON object on stderr and the process exits
// non-zero. Human output is rendered by the caller and stays unchanged when
// JSON mode is off.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// JSON is set by main when the global --json flag is present.
var JSON bool

// EncodeTo writes a single JSON value followed by a newline.
func EncodeTo(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// Emit writes one JSON value to stdout.
func Emit(v any) {
	_ = EncodeTo(os.Stdout, v)
}

// Result emits v as JSON in JSON mode, otherwise calls human.
func Result(v any, human func()) {
	if JSON {
		Emit(v)
		return
	}
	human()
}

// Failf prints an error and exits non-zero.
func Failf(format string, args ...any) {
	FailWith(fmt.Sprintf(format, args...), fmt.Errorf(format, args...))
}

// FailMsg prints a human message or a JSON error and exits non-zero.
func FailMsg(msg string) {
	if JSON {
		_ = EncodeTo(os.Stderr, ErrorResponse{Error: msg})
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}

// FailWith prints a human message or the error as JSON and exits non-zero.
func FailWith(human string, err error) {
	if JSON {
		_ = EncodeTo(os.Stderr, ErrorResponse{Error: err.Error()})
	} else {
		fmt.Fprintln(os.Stderr, human)
	}
	os.Exit(1)
}

// Warn reports a non-fatal warning without changing the exit status.
func Warn(err error) {
	if JSON {
		_ = EncodeTo(os.Stderr, WarningResponse{Warning: err.Error()})
		return
	}
	fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
}

// LogLine is one streamed log line in JSON-lines mode.
type LogLine struct {
	Line string `json:"line"`
}

// JSONLineWriter converts a byte stream into newline-delimited JSON objects.
// It is used for long-running commands such as log streaming.
type JSONLineWriter struct {
	mu  sync.Mutex
	w   io.Writer
	buf []byte
}

// NewJSONLineWriter returns a writer that emits one JSON object per line.
func NewJSONLineWriter(w io.Writer) *JSONLineWriter {
	return &JSONLineWriter{w: w}
}

func (j *JSONLineWriter) Write(p []byte) (int, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.buf = append(j.buf, p...)
	for {
		idx := bytes.IndexByte(j.buf, '\n')
		if idx < 0 {
			break
		}
		line := string(j.buf[:idx])
		j.buf = j.buf[idx+1:]
		if err := EncodeTo(j.w, LogLine{Line: line}); err != nil {
			return len(p), err
		}
	}
	return len(p), nil
}

// Flush emits any buffered partial line.
func (j *JSONLineWriter) Flush() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if len(j.buf) == 0 {
		return nil
	}
	line := string(j.buf)
	j.buf = nil
	return EncodeTo(j.w, LogLine{Line: line})
}
