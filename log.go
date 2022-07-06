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

const (
	tsKey           = "timestamp="
	defaultTSFormat = "2006-01-02T15:04:05.999Z07:00"
)

var (
	hex     = "0123456789abcdef"
	bufPool byteBufferPool
)

// Logger is the interface for all log operations
// related to emitting logs.
type Logger struct {
	mu                   sync.Mutex // Atomic writes.
	out                  io.Writer  // Output destination.
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
		out:                  os.Stderr,
		level:                InfoLevel,
		tsFormat:             defaultTSFormat,
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

	// Get a buffer from the pool.
	buf := bufPool.Get()

	// Write fixed keys to the buffer before writing user provided ones.
	writeTimeToBuf(buf, l.tsFormat, lvl, l.enableColor)
	writeToBuf(buf, "level", lvl, lvl, l.enableColor, true)
	writeStringToBuf(buf, "message", msg, lvl, l.enableColor, true)

	if l.enableCaller {
		writeToBuf(buf, "caller", caller(l.callerSkipFrameCount), lvl, l.enableColor, true)
	}

	// Format the line as logfmt.
	var count int // count is find out if this is the last key in while itering fields.
	for k, v := range fields {
		space := false
		if count != len(fields)-1 {
			space = true
		}
		writeToBuf(buf, k, v, lvl, l.enableColor, space)
		count++
	}
	buf.AppendString("\n")

	l.mu.Lock()
	_, err := l.out.Write(buf.Bytes())
	if err != nil {
		// Should ideally never happen.
		stdlog.Printf("error logging: %v", err)
	}
	l.mu.Unlock()

	buf.Reset()

	// Put the writer back in the pool.
	bufPool.Put(buf)
}

// writeTimeToBuf writes timestamp key + timestamp into buffer.
func writeTimeToBuf(buf *byteBuffer, format string, lvl Level, color bool) {
	if color {
		buf.AppendString(getColoredKey(tsKey, lvl))
	} else {
		buf.AppendString(tsKey)
	}

	buf.AppendTime(time.Now(), format)
	buf.AppendByte(' ')
}

// writeStringToBuf takes key, value and additional options to write to the buffer in logfmt.
func writeStringToBuf(buf *byteBuffer, key string, val string, lvl Level, color, space bool) {
	if color {
		escapeAndWriteString(buf, getColoredKey(key, lvl))
	} else {
		escapeAndWriteString(buf, key)
	}
	buf.AppendByte('=')
	escapeAndWriteString(buf, val)
	if space {
		buf.AppendByte(' ')
	}
}

// writeToBuf takes key, value and additional options to write to the buffer in logfmt.
func writeToBuf(buf *byteBuffer, key string, val any, lvl Level, color, space bool) {
	if color {
		escapeAndWriteString(buf, getColoredKey(key, lvl))
	} else {
		escapeAndWriteString(buf, key)
	}
	buf.AppendByte('=')

	switch v := val.(type) {
	case string:
		escapeAndWriteString(buf, v)
	case int:
		buf.AppendInt(int64(v))
	case int16:
		buf.AppendInt(int64(v))
	case int32:
		buf.AppendInt(int64(v))
	case int64:
		buf.AppendInt(int64(v))
	case float32:
		buf.AppendFloat(float64(v), 32)
	case float64:
		buf.AppendFloat(float64(v), 64)
	case fmt.Stringer:
		escapeAndWriteString(buf, v.String())
	default:
		escapeAndWriteString(buf, fmt.Sprintf("%v", val))
	}

	if space {
		buf.AppendByte(' ')
	}
}

// escapeAndWriteString escapes the string if any unwanted chars are there.
func escapeAndWriteString(buf *byteBuffer, s string) {
	idx := bytes.IndexFunc([]byte(s), checkEscapingRune)
	if idx != -1 {
		writeQuotedString(buf, s)
		return
	}
	buf.AppendString(s)
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

// checkEscapingRune returns true if the rune is to be escaped.
func checkEscapingRune(r rune) bool {
	return r == '=' || r == ' ' || r == '"' || r == utf8.RuneError
}

// writeQuotedString quotes a string before writing to the buffer.
// Taken from: https://github.com/go-logfmt/logfmt/blob/99455b83edb21b32a1f1c0a32f5001b77487b721/jsonstring.go#L95
func writeQuotedString(buf *byteBuffer, s string) {
	buf.AppendByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' {
				i++
				continue
			}
			if start < i {
				buf.AppendString(s[start:i])
			}
			switch b {
			case '\\', '"':
				buf.AppendByte('\\')
				buf.AppendByte(b)
			case '\n':
				buf.AppendByte('\\')
				buf.AppendByte('n')
			case '\r':
				buf.AppendByte('\\')
				buf.AppendByte('r')
			case '\t':
				buf.AppendByte('\\')
				buf.AppendByte('t')
			default:
				// This encodes bytes < 0x20 except for \n, \r, and \t.
				buf.AppendString(`\u00`)
				buf.AppendByte(hex[b>>4])
				buf.AppendByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError {
			if start < i {
				buf.AppendString(s[start:i])
			}
			buf.AppendString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		buf.AppendString(s[start:])
	}
	buf.AppendByte('"')
}
