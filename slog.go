package logf

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// SlogHandler is an implementation of slog.Handler that uses logf for output.
// It allows users of the standard library's slog package to output logs
// in logfmt format using logf's efficient, zero-allocation formatting.
type SlogHandler struct {
	logger      Logger
	attrs       []slog.Attr
	groupPrefix string
}

// NewSlogHandler creates a new slog.Handler that outputs to the given logf.Logger.
// The handler can be used with slog.New() to create a *slog.Logger.
//
// Example:
//
//	logger := logf.New(logf.Opts{})
//	handler := logf.NewSlogHandler(logger)
//	slogger := slog.New(handler)
//	slogger.Info("hello", "key", "value")
func NewSlogHandler(l Logger) *SlogHandler {
	return &SlogHandler{
		logger:      l,
		attrs:       nil,
		groupPrefix: "",
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return slogLevelToLogf(level) >= h.logger.Level
}

// Handle handles the Record. It will only be called when Enabled returns true.
func (h *SlogHandler) Handle(_ context.Context, r slog.Record) error {
	lvl := slogLevelToLogf(r.Level)

	if lvl < h.logger.Level {
		return nil
	}

	buf := bufPool.Get()

	writeTimeToBuf(buf, h.logger.TimestampFormat, lvl, h.logger.EnableColor)
	writeToBuf(buf, "level", lvl, lvl, h.logger.EnableColor, true)
	writeStringToBuf(buf, "message", r.Message, lvl, h.logger.EnableColor, true)

	if h.logger.EnableCaller && r.PC != 0 {
		writeSlogCallerToBuf(buf, "caller", r.PC, lvl, h.logger.EnableColor, true)
	}

	var key string
	for i := range h.logger.DefaultFields {
		if i%2 == 0 {
			key = h.logger.DefaultFields[i].(string)
			continue
		}
		writeToBuf(buf, key, h.logger.DefaultFields[i], lvl, h.logger.EnableColor, true)
	}

	for _, attr := range h.attrs {
		h.writeAttr(buf, attr, lvl, true)
	}

	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(buf, a, lvl, true)
		return true
	})

	n := len(buf.B)
	if n > 0 && buf.B[n-1] == ' ' {
		buf.B[n-1] = '\n'
	} else {
		buf.AppendByte('\n')
	}

	_, err := h.logger.out.Write(buf.Bytes())
	bufPool.Put(buf)

	return err
}

// WithAttrs returns a new Handler whose attributes consist of both the
// receiver's attributes and the arguments.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &SlogHandler{
		logger:      h.logger,
		attrs:       newAttrs,
		groupPrefix: h.groupPrefix,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	var newPrefix string
	if h.groupPrefix == "" {
		newPrefix = name
	} else {
		newPrefix = h.groupPrefix + "." + name
	}

	return &SlogHandler{
		logger:      h.logger,
		attrs:       h.attrs,
		groupPrefix: newPrefix,
	}
}

// writeAttr writes a single slog.Attr to the buffer with proper formatting.
func (h *SlogHandler) writeAttr(buf *byteBuffer, attr slog.Attr, lvl Level, space bool) {
	// Resolve the attribute (handles LogValuer interface).
	attr.Value = attr.Value.Resolve()

	// Skip empty attrs.
	if attr.Equal(slog.Attr{}) {
		return
	}

	key := attr.Key
	if h.groupPrefix != "" {
		key = h.groupPrefix + "." + key
	}

	// Handle the value based on its kind.
	switch attr.Value.Kind() {
	case slog.KindGroup:
		// For groups, recursively write each attribute with the group name as prefix.
		groupAttrs := attr.Value.Group()
		for i, ga := range groupAttrs {
			groupedKey := key + "." + ga.Key
			groupedAttr := slog.Attr{Key: groupedKey, Value: ga.Value}
			// Add space between group attrs, and after the last one if parent needs space.
			isLast := i == len(groupAttrs)-1
			needsSpace := !isLast || space
			h.writeAttrDirect(buf, groupedAttr, lvl, needsSpace)
		}
	default:
		h.writeAttrDirect(buf, slog.Attr{Key: key, Value: attr.Value}, lvl, space)
	}
}

// writeAttrDirect writes a single attribute directly to the buffer.
func (h *SlogHandler) writeAttrDirect(buf *byteBuffer, attr slog.Attr, lvl Level, space bool) {
	switch attr.Value.Kind() {
	case slog.KindString:
		writeStringToBuf(buf, attr.Key, attr.Value.String(), lvl, h.logger.EnableColor, space)
	case slog.KindInt64:
		writeToBuf(buf, attr.Key, attr.Value.Int64(), lvl, h.logger.EnableColor, space)
	case slog.KindUint64:
		writeToBuf(buf, attr.Key, int64(attr.Value.Uint64()), lvl, h.logger.EnableColor, space)
	case slog.KindFloat64:
		writeToBuf(buf, attr.Key, attr.Value.Float64(), lvl, h.logger.EnableColor, space)
	case slog.KindBool:
		writeToBuf(buf, attr.Key, attr.Value.Bool(), lvl, h.logger.EnableColor, space)
	case slog.KindDuration:
		writeDurationToBuf(buf, attr.Key, attr.Value.Duration(), lvl, h.logger.EnableColor, space)
	case slog.KindTime:
		writeTimestampAttr(buf, attr.Key, attr.Value.Time(), h.logger.TimestampFormat, lvl, h.logger.EnableColor, space)
	case slog.KindAny:
		val := attr.Value.Any()
		if err, ok := val.(error); ok {
			writeStringToBuf(buf, attr.Key, err.Error(), lvl, h.logger.EnableColor, space)
		} else {
			writeToBuf(buf, attr.Key, val, lvl, h.logger.EnableColor, space)
		}
	default:
		writeToBuf(buf, attr.Key, attr.Value.Any(), lvl, h.logger.EnableColor, space)
	}
}

// writeTimestampAttr writes a time attribute with proper formatting.
func writeTimestampAttr(buf *byteBuffer, key string, t time.Time, format string, lvl Level, color, space bool) {
	if color {
		escapeAndWriteString(buf, getColoredKey(key, lvl))
	} else {
		escapeAndWriteString(buf, key)
	}
	buf.AppendByte('=')
	buf.AppendTime(t, format)
	if space {
		buf.AppendByte(' ')
	}
}

// writeSlogCallerToBuf writes caller info from slog.Record.PC to buffer.
func writeSlogCallerToBuf(buf *byteBuffer, key string, pc uintptr, lvl Level, color, space bool) {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	if color {
		buf.AppendString(getColoredKey(key, lvl))
	} else {
		buf.AppendString(key)
	}

	buf.AppendByte('=')
	if f.File != "" {
		escapeAndWriteString(buf, f.File)
	} else {
		buf.AppendString("???")
	}
	buf.AppendByte(':')
	buf.AppendInt(int64(f.Line))

	if space {
		buf.AppendByte(' ')
	}
}

// slogLevelToLogf converts a slog.Level to a logf.Level.
func slogLevelToLogf(level slog.Level) Level {
	switch {
	case level < slog.LevelInfo:
		return DebugLevel
	case level < slog.LevelWarn:
		return InfoLevel
	case level < slog.LevelError:
		return WarnLevel
	default:
		return ErrorLevel
	}
}
