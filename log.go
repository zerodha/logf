package logf

import (
	"bytes"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

var hex = "0123456789abcdef"

// Logger is the interface for all log operations
// related to emitting logs.
type Logger struct {
	mu                   sync.Mutex // Atomic writes.
	out                  io.Writer  // Output destination.
	bufW                 sync.Pool  // Buffer pool to accumulate before writing to output.
	level                Level      // Verbosity of logs.
	tsFormat             string     // Timestamp format.
	enableColor          bool       // Colored output.
	enableCaller         bool       // Print caller information.
	callerSkipFrameCount int        // Number of frames to skip when detecting caller
}

// Opts represents various properties
// to configure logger.
type Opts struct {
	Writer          io.Writer
	Lvl             Level
	TimestampFormat string
	EnableColor     bool
	EnableCaller    bool
	// CallerSkipFrameCount is the count of the number of frames to skip when computing the file name and line number
	CallerSkipFrameCount int
}

// Fields is a map of arbitrary KV pairs
// which will be used in logfmt representation of the log.
type Fields map[string]any

// Severity level of the log.
type Level int

const (
	DebugLevel Level = iota // 0
	InfoLevel               // 1
	WarnLevel               // 2
	ErrorLevel              // 3
	FatalLevel              // 4
)

// ANSI escape codes for coloring text in console.
const (
	reset  = "\033[0m"
	purple = "\033[35m"
	red    = "\033[31m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

// Map colors with log level.
var colorLvlMap = [...]string{
	DebugLevel: purple,
	InfoLevel:  cyan,
	WarnLevel:  yellow,
	ErrorLevel: red,
	FatalLevel: red,
}

// New instantiates a logger object.
// It writes to `stderr` as the default and it's non configurable.
func New() *Logger {
	// Initialise logger with sane defaults.
	return &Logger{
		out: os.Stderr,
		bufW: sync.Pool{New: func() any {
			return bytes.NewBuffer([]byte{})
		}},
		level:                InfoLevel,
		tsFormat:             time.RFC3339,
		enableColor:          false,
		enableCaller:         false,
		callerSkipFrameCount: 0,
	}
}

// String representation of the log severity.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "invalid lvl"
	}
}

// SetLevel sets the verbosity for logger.
// Verbosity can be dynamically changed by the caller.
func (l *Logger) SetLevel(lvl Level) {
	l.level = lvl
}

// SetWriter sets the output writer for the logger
func (l *Logger) SetWriter(w io.Writer) {
	l.mu.Lock()
	l.out = w
	l.mu.Unlock()
}

// SetTimestampFormat sets the timestamp format for the `timestamp` key.
func (l *Logger) SetTimestampFormat(f string) {
	l.tsFormat = f
}

// SetColorOutput enables/disables colored output.
func (l *Logger) SetColorOutput(color bool) {
	l.enableColor = color
}

// SetCallerFrame enables/disables the caller source in the log line.
func (l *Logger) SetCallerFrame(caller bool, depth int) {
	l.enableCaller = caller
	l.callerSkipFrameCount = depth
}

// Debug emits a debug log line.
func (l *Logger) Debug(msg string) {
	l.handleLog(msg, DebugLevel, nil)
}

// Info emits a info log line.
func (l *Logger) Info(msg string) {
	l.handleLog(msg, InfoLevel, nil)
}

// Warn emits a warning log line.
func (l *Logger) Warn(msg string) {
	l.handleLog(msg, WarnLevel, nil)
}

// Error emits an error log line.
func (l *Logger) Error(msg string) {
	l.handleLog(msg, ErrorLevel, nil)
}

// Fatal emits a fatal level log line.
// It aborts the current program with an exit code of 1.
func (l *Logger) Fatal(msg string) {
	l.handleLog(msg, FatalLevel, nil)
	os.Exit(1)
}

// WithFields returns a new entry with `fields` set.
func (l *Logger) WithFields(fields Fields) *FieldLogger {
	fl := &FieldLogger{
		fields: fields,
		logger: l,
	}
	return fl
}

// WithError returns a Logger with the "error" key set to `err`.
func (l *Logger) WithError(err error) *FieldLogger {
	if err == nil {
		return &FieldLogger{logger: l}
	}

	return l.WithFields(Fields{
		"error": err.Error(),
	})
}

// handleLog emits the log after filtering log level
// and applying formatting of the fields.
func (l *Logger) handleLog(msg string, lvl Level, fields Fields) {
	// Discard the log if the verbosity is higher.
	// For eg, if the lvl is `3` (error), but the incoming message is `0` (debug), skip it.
	if lvl < l.level {
		return
	}

	now := time.Now().Format(l.tsFormat)

	// Get a buffer from the pool.
	bufW := l.bufW.Get().(*bytes.Buffer)
	defer l.bufW.Put(bufW)

	// Write fixed keys to the buffer before writing user provided ones.
	writeToBuf(bufW, "timestamp", now, lvl, l.enableColor, true)
	writeToBuf(bufW, "level", lvl, lvl, l.enableColor, true)
	writeToBuf(bufW, "message", msg, lvl, l.enableColor, true)

	if l.enableCaller {
		writeToBuf(bufW, "caller", caller(l.callerSkipFrameCount), lvl, l.enableColor, true)
	}

	// Format the line as logfmt.
	var count int // count is find out if this is the last key in while itering fields.
	for k, v := range fields {
		space := false
		if count != len(fields)-1 {
			space = true
		}
		writeToBuf(bufW, k, v, lvl, l.enableColor, space)
		count++
	}
	bufW.WriteString("\n")

	l.mu.Lock()
	_, err := io.Copy(l.out, bufW)
	if err != nil {
		// Should ideally never happen.
		stdlog.Printf("error logging: %v", err)
	}
	l.mu.Unlock()

	bufW.Reset()
}

// writeToBuf takes key, value and additional options to write to the buffer in logfmt.
func writeToBuf(bufW *bytes.Buffer, key string, val any, lvl Level, color, space bool) {
	if color {
		escapeAndWriteString(bufW, getColoredKey(key, lvl))
	} else {
		escapeAndWriteString(bufW, key)
	}
	bufW.WriteByte('=')
	escapeAndWriteString(bufW, getString(val))
	if space {
		bufW.WriteByte(' ')
	}
}

// escapeAndWriteString escapes the string if any unwanted chars are there.
func escapeAndWriteString(bufW *bytes.Buffer, s string) {
	idx := bytes.IndexFunc([]byte(s), checkEscapingRune)
	if idx != -1 {
		writeQuotedString(bufW, s)
		return
	}
	bufW.WriteString(s)
}

// getColoredKey returns a color formatter key based on the log level.
func getColoredKey(k string, lvl Level) string {
	return colorLvlMap[lvl] + k + reset
}

// caller returns the file:line of the caller.
func caller(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 0
	}
	return file + ":" + strconv.Itoa(line)
}

// getString returns a string representation of the given value.
func getString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(int64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'b', 4, 32)
	case float64:
		return strconv.FormatFloat(v, 'b', 4, 64)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// checkEscapingRune returns true if the rune is to be escaped.
func checkEscapingRune(r rune) bool {
	return r == '=' || r == ' ' || r == '"' || r == utf8.RuneError
}

// writeQuotedString quotes a string before writing to the buffer.
// Taken from: https://github.com/go-logfmt/logfmt/blob/99455b83edb21b32a1f1c0a32f5001b77487b721/jsonstring.go#L95
func writeQuotedString(bufW *bytes.Buffer, s string) {
	bufW.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' {
				i++
				continue
			}
			if start < i {
				bufW.WriteString(s[start:i])
			}
			switch b {
			case '\\', '"':
				bufW.WriteByte('\\')
				bufW.WriteByte(b)
			case '\n':
				bufW.WriteByte('\\')
				bufW.WriteByte('n')
			case '\r':
				bufW.WriteByte('\\')
				bufW.WriteByte('r')
			case '\t':
				bufW.WriteByte('\\')
				bufW.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \n, \r, and \t.
				bufW.WriteString(`\u00`)
				bufW.WriteByte(hex[b>>4])
				bufW.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError {
			if start < i {
				bufW.WriteString(s[start:i])
			}
			bufW.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		bufW.WriteString(s[start:])
	}
	bufW.WriteByte('"')
}
