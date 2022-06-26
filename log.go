package logf

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Logger is the interface for all log operations
// related to emitting logs.
type Logger struct {
	mu                   sync.Mutex // Atomic writes.
	out                  io.Writer  // Output destination.
	buf                  []byte     // Accumulate text to write to out.
	level                Level      // Verbosity of logs.
	tsFormat             string     // Timestamp format.
	enableColor          bool       // Colored output.
	enableCaller         bool       // Print caller information.
	callerSkipFrameCount int        // Number of frames to skip when detecting caller
	fields               Fields     // Arbitrary map of KV pair to log.
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

// 0 - debug
// 1 - info
// 2 - warn
// 3 - error
// 4 - fatal
const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// New instantiates a logger object.
// It writes to `stderr` as the default and it's non configurable.
func New() *Logger {
	// Initialise logger with sane defaults.
	return &Logger{
		out:                  os.Stderr,
		level:                InfoLevel,
		tsFormat:             time.RFC3339,
		enableColor:          true,
		enableCaller:         false,
		callerSkipFrameCount: 0,
		fields:               make(Fields, 0),
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
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = lvl
}

// SetWriter sets the output writer for the logger
func (l *Logger) SetWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// SetTimestampFormat sets the timestamp format for the `timestamp` key.
func (l *Logger) SetTimestampFormat(f string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tsFormat = f
}

// SetColorOutput enables/disables colored output.
func (l *Logger) SetColorOutput(color bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enableColor = color
}

// SetCallerFrame enables/disables the caller source in the log line.
func (l *Logger) SetCallerFrame(caller bool, depth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enableCaller = caller
	l.callerSkipFrameCount = depth
}

// Debug emits a debug log line.
func (l *Logger) Debug(msg string) {
	l.handleLog(msg, DebugLevel)
}

// Info emits a info log line.
func (l *Logger) Info(msg string) {
	l.handleLog(msg, InfoLevel)
}

// Warn emits a warning log line.
func (l *Logger) Warn(msg string) {
	l.handleLog(msg, WarnLevel)
}

// Error emits an error log line.
func (l *Logger) Error(msg string) {
	l.handleLog(msg, ErrorLevel)
}

// Fatal emits a fatal level log line.
// It aborts the current program with an exit code of 1.
func (l *Logger) Fatal(msg string) {
	l.handleLog(msg, FatalLevel)
	os.Exit(1)
}

// WithFields returns a new entry with `fields` set.
func (l *Logger) WithFields(fields Fields) *Logger {
	if fields == nil {
		fields = make(Fields, 0)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Set the fields in logger.
	l.fields = fields

	return l
}

// WithError returns a Logger with the "error" key set to `err`.
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}

	return l.WithFields(Fields{
		"error": err.Error(),
	})
}

// handleLog emits the log after filtering log level
// and applying formatting of the fields.
func (l *Logger) handleLog(msg string, lvl Level) {
	// Lock the map to prevet concurrent access to fields map.
	l.mu.Lock()
	defer l.mu.Unlock()

	// Discard the log if the verbosity is higher.
	// For eg, if the lvl is `3` (error), but the incoming message is `0` (debug), skip it.
	if lvl < l.level {
		return
	}

	// Create the output map and copy fields.
	outMap := make(map[string]any)
	// Copy from the original map to the target map
	for k, v := range l.fields {
		outMap[k] = v
	}

	// Collect all user defined keys and sort them.
	var keys []string
	for k := range outMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// All the extra keys will be appended to it after these fixed keys.
	fixedKeys := []string{"timestamp", "level", "message"}
	fixedKeys = append(fixedKeys, keys...)
	if l.enableCaller {
		fixedKeys = append(fixedKeys, "caller")
	}

	// Add extra metadata.
	outMap["message"] = msg
	outMap["level"] = lvl
	outMap["timestamp"] = time.Now().Format(l.tsFormat)
	if l.enableCaller {
		// Release the lock because getting caller info is expensive.
		l.mu.Unlock()
		outMap["caller"] = caller(l.callerSkipFrameCount)
		l.mu.Lock()
	}

	// Format the line as logfmt.
	var line string
	for _, k := range fixedKeys {
		if l.enableColor {
			// Release the lock because coloring the key is expensive.
			l.mu.Unlock()
			line += fmt.Sprintf("%s=%v ", getColoredKey(k, lvl.String()), outMap[k])
			l.mu.Lock()
		} else {
			line += fmt.Sprintf("%s=%v ", k, outMap[k])
		}
	}

	dest := strings.TrimRight(line, " ")

	// Reset buffer.
	l.buf = l.buf[:0]
	l.buf = append(l.buf, dest...)
	if len(dest) == 0 || dest[len(dest)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}

	_, err := l.out.Write(l.buf)
	if err != nil {
		// Should ideally never happen.
		stdlog.Printf("error logging: %v", err)
	}
}

// getColoredKey returns a color formatter key based on the log level.
func getColoredKey(k string, lvl string) string {
	var (
		white  = color.New(color.FgWhite, color.Bold).SprintFunc()
		cyan   = color.New(color.FgCyan, color.Bold).SprintFunc()
		red    = color.New(color.FgRed, color.Bold).SprintFunc()
		yellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	)

	switch lvl {
	default:
		return k
	case "debug":
		return white(k)
	case "info":
		return cyan(k)
	case "warn":
		return yellow(k)
	case "fatal", "error":
		return red(k)
	}
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
