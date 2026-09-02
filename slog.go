package logf

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// SlogHandler is an implementation of slog.Handler that uses logf for output.
// It allows users of the standard library's slog package to output logs in
// logfmt format using logf's efficient formatting.
type SlogHandler struct {
	logger               Logger
	attrs                []slogHandlerAttr
	preformatted         []byte
	defaultFields        []any
	preformattedDefaults []byte
	groupPrefix          string
	sourceCache          *sync.Map
}

type slogHandlerAttr struct {
	attr   slog.Attr
	prefix string
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
	handler := &SlogHandler{
		logger:        l,
		defaultFields: append([]any(nil), l.DefaultFields...),
	}
	if !l.EnableColor {
		buf := &byteBuffer{}
		handler.appendDefaultFields(buf, InfoLevel)
		handler.preformattedDefaults = buf.B
	}
	if l.EnableCaller {
		handler.sourceCache = &sync.Map{}
	}
	return handler
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

	writeSlogTimeToBuf(buf, r.Time, h.logger.TimestampFormat, lvl, h.logger.EnableColor)
	writeToBuf(buf, "level", lvl, lvl, h.logger.EnableColor, true)
	writeStringToBuf(buf, "message", r.Message, lvl, h.logger.EnableColor, true)

	if h.logger.EnableCaller && r.PC != 0 {
		h.writeCallerToBuf(buf, "caller", r.PC, lvl, h.logger.EnableColor, true)
	}

	if h.logger.EnableColor {
		h.appendDefaultFields(buf, lvl)
	} else {
		buf.B = append(buf.B, h.preformattedDefaults...)
	}

	if h.logger.EnableColor {
		for _, attr := range h.attrs {
			h.appendAttr(buf, attr.attr, lvl, attr.prefix, nil)
		}
	} else {
		buf.B = append(buf.B, h.preformatted...)
	}
	r.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(buf, attr, lvl, h.groupPrefix, nil)
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
	if len(attrs) == 0 {
		return h
	}

	handlerAttrs := make([]slogHandlerAttr, len(h.attrs)+len(attrs))
	copy(handlerAttrs, h.attrs)
	for i, attr := range attrs {
		handlerAttrs[len(h.attrs)+i] = slogHandlerAttr{attr: resolveSlogAttr(attr), prefix: h.groupPrefix}
	}

	preformatted := append([]byte(nil), h.preformatted...)
	if !h.logger.EnableColor {
		buf := &byteBuffer{B: preformatted}
		for _, attr := range handlerAttrs[len(h.attrs):] {
			h.appendAttr(buf, attr.attr, InfoLevel, attr.prefix, nil)
		}
		preformatted = buf.B
	}

	return &SlogHandler{
		logger:               h.logger,
		attrs:                handlerAttrs,
		preformatted:         preformatted,
		defaultFields:        h.defaultFields,
		preformattedDefaults: h.preformattedDefaults,
		groupPrefix:          h.groupPrefix,
		sourceCache:          h.sourceCache,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &SlogHandler{
		logger:               h.logger,
		attrs:                h.attrs,
		preformatted:         h.preformatted,
		defaultFields:        h.defaultFields,
		preformattedDefaults: h.preformattedDefaults,
		groupPrefix:          slogKey(h.groupPrefix, name),
		sourceCache:          h.sourceCache,
	}
}

func (h *SlogHandler) appendDefaultFields(buf *byteBuffer, lvl Level) {
	var key string
	for i := range h.defaultFields {
		if i%2 == 0 {
			key = h.defaultFields[i].(string)
			continue
		}
		writeToBuf(buf, key, h.defaultFields[i], lvl, h.logger.EnableColor, true)
	}
}

// resolveSlogAttr resolves a persistent attribute, including nested groups.
func resolveSlogAttr(attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	if attr.Value.Kind() != slog.KindGroup {
		return attr
	}

	group := attr.Value.Group()

	resolved := make([]slog.Attr, len(group))
	for i, groupAttr := range group {
		resolved[i] = resolveSlogAttr(groupAttr)
	}
	attr.Value = slog.GroupValue(resolved...)
	return attr
}

type slogGroup struct {
	parent *slogGroup
	name   string
}

// appendAttr appends attr and nested groups in logfmt form.
func (h *SlogHandler) appendAttr(buf *byteBuffer, attr slog.Attr, lvl Level, prefix string, group *slogGroup) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			group = &slogGroup{parent: group, name: attr.Key}
		}
		for _, groupAttr := range attr.Value.Group() {
			h.appendAttr(buf, groupAttr, lvl, prefix, group)
		}
		return
	}

	writeSlogAttr(buf, prefix, group, attr.Key, attr.Value, lvl, h.logger.EnableColor)
}

func slogKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	if key == "" {
		return prefix
	}
	return prefix + "." + key
}

// writeSlogAttr writes one resolved, non-group attribute.
func writeSlogAttr(buf *byteBuffer, prefix string, group *slogGroup, key string, value slog.Value, lvl Level, color bool) {
	appendSlogKey(buf, prefix, group, key, lvl, color)
	buf.AppendByte('=')

	switch value.Kind() {
	case slog.KindString:
		escapeAndWriteString(buf, value.String())
	case slog.KindInt64:
		buf.AppendInt(value.Int64())
	case slog.KindUint64:
		buf.AppendUint(value.Uint64())
	case slog.KindFloat64:
		buf.AppendFloat(value.Float64(), 64)
	case slog.KindBool:
		buf.AppendBool(value.Bool())
	case slog.KindDuration:
		buf.AppendDuration(value.Duration())
	case slog.KindTime:
		buf.AppendTime(value.Time(), time.RFC3339Nano)
	case slog.KindAny:
		appendValueToBuf(buf, value.Any())
	default:
		appendValueToBuf(buf, value.Any())
	}
	buf.AppendByte(' ')
}

func appendSlogKey(buf *byteBuffer, prefix string, group *slogGroup, key string, lvl Level, color bool) {
	quoted := slogKeyNeedsQuoting(prefix, group, key)
	if quoted {
		buf.AppendByte('"')
	}
	if color {
		buf.AppendString(colorLvlMap[lvl])
	}

	wrote := false
	appendSlogKeyPart(buf, prefix, quoted, &wrote)
	appendSlogGroup(buf, group, quoted, &wrote)
	appendSlogKeyPart(buf, key, quoted, &wrote)

	if color {
		buf.AppendString(reset)
	}
	if quoted {
		buf.AppendByte('"')
	}
}

func appendSlogGroup(buf *byteBuffer, group *slogGroup, quoted bool, wrote *bool) {
	if group == nil {
		return
	}
	appendSlogGroup(buf, group.parent, quoted, wrote)
	appendSlogKeyPart(buf, group.name, quoted, wrote)
}

func appendSlogKeyPart(buf *byteBuffer, part string, quoted bool, wrote *bool) {
	if part == "" {
		return
	}
	if *wrote {
		buf.AppendByte('.')
	}
	if quoted {
		appendQuotedStringContent(buf, part)
	} else {
		buf.AppendString(part)
	}
	*wrote = true
}

func slogKeyNeedsQuoting(prefix string, group *slogGroup, key string) bool {
	parts := 0
	if prefix != "" {
		parts++
		if needsEscaping(prefix) {
			return true
		}
	}
	for current := group; current != nil; current = current.parent {
		parts++
		if needsEscaping(current.name) {
			return true
		}
	}
	if key != "" {
		parts++
		if needsEscaping(key) {
			return true
		}
	}
	if parts != 1 {
		return false
	}
	if prefix != "" {
		return prefix == "null"
	}
	if group != nil {
		return group.name == "null"
	}
	return key == "null"
}

func writeSlogTimeToBuf(buf *byteBuffer, t time.Time, format string, lvl Level, color bool) {
	if color {
		buf.AppendString(getColoredKey(tsKey, lvl))
	} else {
		buf.AppendString(tsKey)
	}
	buf.AppendTime(t, format)
	buf.AppendByte(' ')
}

type slogSource struct {
	function string
	file     string
	line     int
}

func sourceForPC(cache *sync.Map, pc uintptr) slogSource {
	if cached, ok := cache.Load(pc); ok {
		return cached.(slogSource)
	}

	fs := runtime.CallersFrames([]uintptr{pc})
	frame, _ := fs.Next()
	source := slogSource{function: frame.Function, file: frame.File, line: frame.Line}
	cached, _ := cache.LoadOrStore(pc, source)
	return cached.(slogSource)
}

// writeCallerToBuf writes caller info from slog.Record.PC to buf. Source
// resolution is cached because a logger normally sees the same callsite PCs.
func (h *SlogHandler) writeCallerToBuf(buf *byteBuffer, key string, pc uintptr, lvl Level, color, space bool) {
	source := sourceForPC(h.sourceCache, pc)

	if color {
		buf.AppendString(getColoredKey(key, lvl))
	} else {
		buf.AppendString(key)
	}

	buf.AppendByte('=')
	if source.file != "" {
		escapeAndWriteString(buf, source.file)
	} else {
		buf.AppendString("???")
	}
	buf.AppendByte(':')
	buf.AppendInt(int64(source.line))

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
