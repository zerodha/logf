package logf

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"runtime"
	"strings"
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
	exit    = func() { os.Exit(1) }
)

type Opts struct {
	Writer               io.Writer
	Level                Level
	TimestampFormat      string
	EnableColor          bool
	EnableCaller         bool
	CallerSkipFrameCount int

	// These fields will be printed with every log.
	DefaultFields []interface{}
}

// Logger is the interface for all log operations
// related to emitting logs.
type Logger struct {
	out io.Writer // Output destination.
	Opts
}

// Severity level of the log.
type Level int

const (
	DebugLevel Level = iota + 1 // 1
	InfoLevel                   // 2
	WarnLevel                   // 3
	ErrorLevel                  // 4
	FatalLevel                  // 5
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
func New(opts Opts) Logger {
	// Initialize fallbacks if unspecified by user.
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}
	if opts.TimestampFormat == "" {
		opts.TimestampFormat = defaultTSFormat
	}
	if opts.Level == 0 {
		opts.Level = InfoLevel
	}
	if opts.CallerSkipFrameCount == 0 {
		opts.CallerSkipFrameCount = 3
	}

	if len(opts.DefaultFields)%2 != 0 {
		opts.DefaultFields = opts.DefaultFields[0 : len(opts.DefaultFields)-1]
	}

	return Logger{
		out:  newSyncWriter(opts.Writer),
		Opts: opts,
	}
}

// syncWriter is a wrapper around io.Writer that
// synchronizes writes using a mutex.
type syncWriter struct {
	sync.Mutex
	w io.Writer
}

// Write synchronously to the underlying io.Writer.
func (w *syncWriter) Write(p []byte) (int, error) {
	w.Lock()
	n, err := w.w.Write(p)
	w.Unlock()
	return n, err
}

// newSyncWriter wraps an io.Writer with syncWriter. It can
// be used as an io.Writer as syncWriter satisfies the io.Writer interface.
func newSyncWriter(in io.Writer) *syncWriter {
	if in == nil {
		return &syncWriter{w: os.Stderr}
	}

	return &syncWriter{w: in}
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

func LevelFromString(lvl string) (Level, error) {
	switch lvl {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return 0, fmt.Errorf("invalid level")
	}
}

// Debug emits a debug log line.
func (l Logger) Debug(msg string, fields ...interface{}) {
	l.handleLog(msg, DebugLevel, fields...)
}

// Info emits a info log line.
func (l Logger) Info(msg string, fields ...interface{}) {
	l.handleLog(msg, InfoLevel, fields...)
}

// Warn emits a warning log line.
func (l Logger) Warn(msg string, fields ...interface{}) {
	l.handleLog(msg, WarnLevel, fields...)
}

// Error emits an error log line.
func (l Logger) Error(msg string, fields ...interface{}) {
	l.handleLog(msg, ErrorLevel, fields...)
}

// Fatal emits a fatal level log line.
// It aborts the current program with an exit code of 1.
func (l Logger) Fatal(msg string, fields ...interface{}) {
	l.handleLog(msg, FatalLevel, fields...)
	exit()
}

// handleLog emits the log after filtering log level
// and applying formatting of the fields.
func (l Logger) handleLog(msg string, lvl Level, fields ...interface{}) {
	// Discard the log if the verbosity is higher.
	// For eg, if the lvl is `3` (error), but the incoming message is `0` (debug), skip it.
	if lvl < l.Opts.Level {
		return
	}

	// Get a buffer from the pool.
	buf := bufPool.Get()

	// Write fixed keys to the buffer before writing user provided ones.
	writeTimeToBuf(buf, l.Opts.TimestampFormat, lvl, l.Opts.EnableColor)
	writeToBuf(buf, "level", lvl, lvl, l.Opts.EnableColor, true)
	writeStringToBuf(buf, "message", msg, lvl, l.Opts.EnableColor, true)

	if l.Opts.EnableCaller {
		writeCallerToBuf(buf, "caller", l.Opts.CallerSkipFrameCount, lvl, l.EnableColor, true)
	}

	// Format the line as logfmt.
	var (
		count      int // to find out if this is the last key in while itering fields.
		fieldCount = len(l.DefaultFields) + len(fields)
		key        string
		val        interface{}
	)

	// If there are odd number of fields, ignore the last.
	if fieldCount%2 != 0 {
		fields = fields[0 : len(fields)-1]
	}

	for i := range l.DefaultFields {
		space := false
		if count != fieldCount-1 {
			space = true
		}

		if i%2 == 0 {
			key = l.DefaultFields[i].(string)
			continue
		} else {
			val = l.DefaultFields[i]
		}

		writeToBuf(buf, key, val, lvl, l.Opts.EnableColor, space)
		count++
	}

	for i := range fields {
		space := false
		if count != fieldCount-1 {
			space = true
		}

		if i%2 == 0 {
			key = fields[i].(string)
			continue
		} else {
			val = fields[i]
		}

		writeToBuf(buf, key, val, lvl, l.Opts.EnableColor, space)
		count++
	}
	buf.AppendString("\n")

	_, err := l.out.Write(buf.Bytes())
	if err != nil {
		// Should ideally never happen.
		stdlog.Printf("error logging: %v", err)
	}

	// Put the writer back in the pool. It resets the underlying byte buffer.
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
func writeStringToBuf(buf *byteBuffer, key, val string, lvl Level, color, space bool) {
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

func writeCallerToBuf(buf *byteBuffer, key string, depth int, lvl Level, color, space bool) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 0
	}
	if color {
		buf.AppendString(getColoredKey(key, lvl))
	} else {
		buf.AppendString(key)
	}
	buf.AppendByte('=')
	escapeAndWriteString(buf, file)
	buf.AppendByte(':')
	buf.AppendInt(int64(line))
	if space {
		buf.AppendByte(' ')
	}
}

// writeToBuf takes key, value and additional options to write to the buffer in logfmt.
func writeToBuf(buf *byteBuffer, key string, val interface{}, lvl Level, color, space bool) {
	if color {
		escapeAndWriteString(buf, getColoredKey(key, lvl))
	} else {
		escapeAndWriteString(buf, key)
	}
	buf.AppendByte('=')

	switch v := val.(type) {
	case nil:
		buf.AppendString("null")
	case []byte:
		escapeAndWriteString(buf, string(v))
	case string:
		escapeAndWriteString(buf, v)
	case int:
		buf.AppendInt(int64(v))
	case int8:
		buf.AppendInt(int64(v))
	case int16:
		buf.AppendInt(int64(v))
	case int32:
		buf.AppendInt(int64(v))
	case int64:
		buf.AppendInt(v)
	case float32:
		buf.AppendFloat(float64(v), 32)
	case float64:
		buf.AppendFloat(v, 64)
	case bool:
		buf.AppendBool(v)
	case error:
		escapeAndWriteString(buf, v.Error())
	case fmt.Stringer:
		escapeAndWriteString(buf, v.String())
	default:
		escapeAndWriteString(buf, fmt.Sprintf("%v", val))
	}

	if space {
		buf.AppendByte(' ')
	}
}

// escapeAndWriteString escapes the string if interface{} unwanted chars are there.
func escapeAndWriteString(buf *byteBuffer, s string) {
	idx := strings.IndexFunc(s, checkEscapingRune)
	if idx != -1 || s == "null" {
		writeQuotedString(buf, s)
		return
	}
	buf.AppendString(s)
}

// getColoredKey returns a color formatter key based on the log level.
func getColoredKey(k string, lvl Level) string {
	return colorLvlMap[lvl] + k + reset
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
