package logf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

// SlogJSONHandler is a slog.Handler that writes line-delimited JSON objects.
// It follows slog.JSONHandler's built-in field names and value encoding while
// using the log level, output writer, default fields, and caller setting from
// Logger.
type SlogJSONHandler struct {
	logger               Logger
	groups               []string
	nOpenGroups          int
	preformatted         []byte
	preformattedDefaults []byte
	sourceCache          *sync.Map
}

// NewSlogJSONHandler creates a JSON slog handler backed by l.
func NewSlogJSONHandler(l Logger) *SlogJSONHandler {
	handler := &SlogJSONHandler{logger: l}
	if l.EnableCaller {
		handler.sourceCache = &sync.Map{}
	}

	state := jsonHandleState{buf: &byteBuffer{}}
	var key string
	for i := range l.DefaultFields {
		if i%2 == 0 {
			key = l.DefaultFields[i].(string)
			continue
		}
		state.appendKey(key)
		appendJSONAny(state.buf, l.DefaultFields[i])
	}
	handler.preformattedDefaults = state.buf.B
	return handler
}

// Enabled reports whether the handler handles records at level.
func (h *SlogJSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return slogLevelToLogf(level) >= h.logger.Level
}

// Handle writes r as one line-delimited JSON object.
func (h *SlogJSONHandler) Handle(_ context.Context, r slog.Record) error {
	if slogLevelToLogf(r.Level) < h.logger.Level {
		return nil
	}

	buf := bufPool.Get()
	state := jsonHandleState{buf: buf}
	buf.AppendByte('{')

	if !r.Time.IsZero() {
		state.appendKey(slog.TimeKey)
		appendJSONTime(buf, r.Time)
	}
	state.appendKey(slog.LevelKey)
	appendJSONString(buf, r.Level.String())
	if h.logger.EnableCaller && r.PC != 0 {
		state.appendKey(slog.SourceKey)
		appendJSONSource(buf, sourceForPC(h.sourceCache, r.PC))
	}
	state.appendKey(slog.MessageKey)
	appendJSONString(buf, r.Message)

	if len(h.preformattedDefaults) > 0 {
		buf.AppendByte(',')
		buf.B = append(buf.B, h.preformattedDefaults...)
		state.comma = true
	}
	if len(h.preformatted) > 0 {
		buf.AppendByte(',')
		buf.B = append(buf.B, h.preformatted...)
		state.comma = true
	}

	openGroups := h.nOpenGroups
	if r.NumAttrs() > 0 {
		pos, comma := len(buf.B), state.comma
		for _, group := range h.groups[h.nOpenGroups:] {
			state.openGroup(group)
		}

		wrote := false
		r.Attrs(func(attr slog.Attr) bool {
			if state.appendAttr(attr) {
				wrote = true
			}
			return true
		})
		if wrote {
			openGroups = len(h.groups)
		} else {
			buf.B = buf.B[:pos]
			state.comma = comma
		}
	}

	for range openGroups {
		buf.AppendByte('}')
	}
	buf.AppendByte('}')
	buf.AppendByte('\n')

	_, err := h.logger.out.Write(buf.Bytes())
	bufPool.Put(buf)
	return err
}

// WithAttrs returns a new handler with attrs preformatted once.
func (h *SlogJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	preformatted := append([]byte(nil), h.preformatted...)
	state := jsonHandleState{
		buf:   &byteBuffer{B: preformatted},
		comma: len(preformatted) > 0,
	}
	pos, comma := len(preformatted), state.comma
	for _, group := range h.groups[h.nOpenGroups:] {
		state.openGroup(group)
	}
	if !state.appendAttrs(attrs) {
		state.buf.B = state.buf.B[:pos]
		state.comma = comma
		return h
	}

	return &SlogJSONHandler{
		logger:               h.logger,
		groups:               h.groups,
		nOpenGroups:          len(h.groups),
		preformatted:         state.buf.B,
		preformattedDefaults: h.preformattedDefaults,
		sourceCache:          h.sourceCache,
	}
}

// WithGroup returns a new handler with name appended to its group path.
func (h *SlogJSONHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	groups := make([]string, len(h.groups), len(h.groups)+1)
	copy(groups, h.groups)
	groups = append(groups, name)
	return &SlogJSONHandler{
		logger:               h.logger,
		groups:               groups,
		nOpenGroups:          h.nOpenGroups,
		preformatted:         h.preformatted,
		preformattedDefaults: h.preformattedDefaults,
		sourceCache:          h.sourceCache,
	}
}

type jsonHandleState struct {
	buf   *byteBuffer
	comma bool
}

func (s *jsonHandleState) appendKey(key string) {
	if s.comma {
		s.buf.AppendByte(',')
	}
	appendJSONString(s.buf, key)
	s.buf.AppendByte(':')
	s.comma = true
}

func (s *jsonHandleState) openGroup(name string) {
	s.appendKey(name)
	s.buf.AppendByte('{')
	s.comma = false
}

func (s *jsonHandleState) appendAttrs(attrs []slog.Attr) bool {
	wrote := false
	for _, attr := range attrs {
		if s.appendAttr(attr) {
			wrote = true
		}
	}
	return wrote
}

func (s *jsonHandleState) appendAttr(attr slog.Attr) bool {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return false
	}
	if attr.Value.Kind() != slog.KindGroup {
		s.appendKey(attr.Key)
		appendJSONValue(s.buf, attr.Value)
		return true
	}

	attrs := attr.Value.Group()
	if len(attrs) == 0 {
		return false
	}
	pos, comma := len(s.buf.B), s.comma
	if attr.Key != "" {
		s.openGroup(attr.Key)
	}
	if !s.appendAttrs(attrs) {
		s.buf.B = s.buf.B[:pos]
		s.comma = comma
		return false
	}
	if attr.Key != "" {
		s.buf.AppendByte('}')
		s.comma = true
	}
	return true
}

func appendJSONValue(buf *byteBuffer, value slog.Value) {
	switch value.Kind() {
	case slog.KindString:
		appendJSONString(buf, value.String())
	case slog.KindInt64:
		buf.AppendInt(value.Int64())
	case slog.KindUint64:
		buf.AppendUint(value.Uint64())
	case slog.KindFloat64:
		appendJSONFloat(buf, value.Float64())
	case slog.KindBool:
		buf.AppendBool(value.Bool())
	case slog.KindDuration:
		buf.AppendInt(int64(value.Duration()))
	case slog.KindTime:
		appendJSONTime(buf, value.Time())
	case slog.KindAny:
		appendJSONAny(buf, value.Any())
	default:
		panic("logf: invalid slog value kind")
	}
}

func appendJSONFloat(buf *byteBuffer, value float64) {
	if math.IsInf(value, 0) || math.IsNaN(value) {
		buf.AppendByte('"')
		buf.B = appendEscapedJSONString(buf.B, "!ERROR:json: unsupported value: ")
		buf.B = strconv.AppendFloat(buf.B, value, 'g', -1, 64)
		buf.AppendByte('"')
		return
	}

	format := byte('f')
	absolute := math.Abs(value)
	if absolute != 0 && (absolute < 1e-6 || absolute >= 1e21) {
		format = 'e'
	}
	buf.B = strconv.AppendFloat(buf.B, value, format, -1, 64)
	if format == 'e' {
		n := len(buf.B)
		if n >= 4 && buf.B[n-4] == 'e' && buf.B[n-3] == '-' && buf.B[n-2] == '0' {
			buf.B[n-2] = buf.B[n-1]
			buf.B = buf.B[:n-1]
		}
	}
}

func appendJSONTime(buf *byteBuffer, value time.Time) {
	if year := value.Year(); year < 0 || year >= 10000 {
		appendJSONError(buf, "time.Time year outside of range [0,9999]")
		return
	}
	buf.AppendByte('"')
	buf.AppendTime(value, time.RFC3339Nano)
	buf.AppendByte('"')
}

func appendJSONSource(buf *byteBuffer, source slogSource) {
	state := jsonHandleState{buf: buf}
	buf.AppendByte('{')
	if source.function != "" {
		state.appendKey("function")
		appendJSONString(buf, source.function)
	}
	if source.file != "" {
		state.appendKey("file")
		appendJSONString(buf, source.file)
	}
	if source.line != 0 {
		state.appendKey("line")
		buf.AppendInt(int64(source.line))
	}
	buf.AppendByte('}')
}

func appendJSONString(buf *byteBuffer, value string) {
	buf.AppendByte('"')
	buf.B = appendEscapedJSONString(buf.B, value)
	buf.AppendByte('"')
}

func appendJSONError(buf *byteBuffer, message string) {
	buf.AppendByte('"')
	buf.B = appendEscapedJSONString(buf.B, "!ERROR:")
	buf.B = appendEscapedJSONString(buf.B, message)
	buf.AppendByte('"')
}

func appendJSONErrorValue(buf *byteBuffer, err error) {
	appendJSONString(buf, fmt.Sprintf("!ERROR:%v", err))
}

func appendEscapedJSONString(buf []byte, value string) []byte {
	start := 0
	for i := 0; i < len(value); {
		if b := value[i]; b < utf8.RuneSelf {
			if b >= 0x20 && b != '\\' && b != '"' {
				i++
				continue
			}
			buf = append(buf, value[start:i]...)
			buf = append(buf, '\\')
			switch b {
			case '\\', '"':
				buf = append(buf, b)
			case '\n':
				buf = append(buf, 'n')
			case '\r':
				buf = append(buf, 'r')
			case '\t':
				buf = append(buf, 't')
			default:
				buf = append(buf, 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}

		r, size := utf8.DecodeRuneInString(value[i:])
		if r == utf8.RuneError && size == 1 {
			buf = append(buf, value[start:i]...)
			buf = append(buf, `\ufffd`...)
			i++
			start = i
			continue
		}
		if r == '\u2028' || r == '\u2029' {
			buf = append(buf, value[start:i]...)
			buf = append(buf, `\u202`...)
			buf = append(buf, hex[r&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	return append(buf, value[start:]...)
}

type pooledJSONEncoder struct {
	buf     bytes.Buffer
	encoder *json.Encoder
}

var slogJSONEncoderPool = sync.Pool{New: func() any {
	pooled := &pooledJSONEncoder{}
	pooled.encoder = json.NewEncoder(&pooled.buf)
	pooled.encoder.SetEscapeHTML(false)
	return pooled
}}

func appendJSONAny(buf *byteBuffer, value any) {
	start := len(buf.B)
	defer func() {
		if recovered := recover(); recovered != nil {
			buf.B = buf.B[:start]
			reflected := reflect.ValueOf(value)
			if reflected.Kind() == reflect.Pointer && reflected.IsNil() {
				appendJSONString(buf, "<nil>")
				return
			}
			appendJSONString(buf, fmt.Sprintf("!PANIC: %v", recovered))
		}
	}()
	appendJSONAnyValue(buf, value)
}

func appendJSONAnyValue(buf *byteBuffer, value any) {
	if err, ok := value.(error); ok {
		if _, marshaler := value.(json.Marshaler); !marshaler {
			appendJSONString(buf, err.Error())
			return
		}
	}

	pooled := slogJSONEncoderPool.Get().(*pooledJSONEncoder)
	defer func() {
		if pooled.buf.Cap() <= 16*1024 {
			pooled.buf.Reset()
			slogJSONEncoderPool.Put(pooled)
		}
	}()

	if err := pooled.encoder.Encode(value); err != nil {
		appendJSONErrorValue(buf, err)
		return
	}
	encoded := pooled.buf.Bytes()
	buf.B = append(buf.B, encoded[:len(encoded)-1]...)
}
