package logf

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlogHandler_BasicLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	slogger := slog.New(handler)

	slogger.Info("test message", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "level=info")
	assert.Contains(t, output, `message="test message"`)
	assert.Contains(t, output, "key=value")
}

func TestSlogHandler_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           WarnLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	slogger := slog.New(handler)

	// These should be filtered out.
	slogger.Debug("debug message")
	slogger.Info("info message")

	// This should appear.
	slogger.Warn("warn message")
	slogger.Error("error message")

	output := buf.String()
	assert.NotContains(t, output, "debug message")
	assert.NotContains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestSlogHandler_Enabled(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  Level
		slogLevel    slog.Level
		wantEnabled  bool
	}{
		{"debug logger, debug msg", DebugLevel, slog.LevelDebug, true},
		{"debug logger, info msg", DebugLevel, slog.LevelInfo, true},
		{"info logger, debug msg", InfoLevel, slog.LevelDebug, false},
		{"info logger, info msg", InfoLevel, slog.LevelInfo, true},
		{"warn logger, info msg", WarnLevel, slog.LevelInfo, false},
		{"warn logger, warn msg", WarnLevel, slog.LevelWarn, true},
		{"error logger, warn msg", ErrorLevel, slog.LevelWarn, false},
		{"error logger, error msg", ErrorLevel, slog.LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(Opts{Level: tt.loggerLevel})
			handler := NewSlogHandler(logger)
			assert.Equal(t, tt.wantEnabled, handler.Enabled(nil, tt.slogLevel))
		})
	}
}

func TestSlogHandler_WithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	handlerWithAttrs := handler.WithAttrs([]slog.Attr{
		slog.String("component", "test"),
		slog.Int("version", 1),
	})
	slogger := slog.New(handlerWithAttrs)

	slogger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "component=test")
	assert.Contains(t, output, "version=1")
}

func TestSlogHandler_WithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	handlerWithGroup := handler.WithGroup("request")
	slogger := slog.New(handlerWithGroup)

	slogger.Info("test message", "id", "123")

	output := buf.String()
	assert.Contains(t, output, "request.id=123")
}

func TestSlogHandler_NestedGroups(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	handlerWithGroups := handler.WithGroup("http").WithGroup("request")
	slogger := slog.New(handlerWithGroups)

	slogger.Info("test message", "method", "GET")

	output := buf.String()
	assert.Contains(t, output, "http.request.method=GET")
}

func TestSlogHandler_DifferentTypes(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	slogger := slog.New(handler)

	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	testDuration := 5 * time.Second

	slogger.Info("test message",
		"string", "hello",
		"int", 42,
		"float", 3.14,
		"bool", true,
		"duration", testDuration,
		"time", testTime,
	)

	output := buf.String()
	assert.Contains(t, output, "string=hello")
	assert.Contains(t, output, "int=42")
	assert.Contains(t, output, "float=3.14")
	assert.Contains(t, output, "bool=true")
	assert.Contains(t, output, "duration=5s")
	assert.Contains(t, output, "time=2024-01-15")
}

func TestSlogHandler_ErrorType(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	slogger := slog.New(handler)

	testErr := errors.New("something went wrong")
	slogger.Error("operation failed", "error", testErr)

	output := buf.String()
	assert.Contains(t, output, "level=error")
	assert.Contains(t, output, `error="something went wrong"`)
}

func TestSlogHandler_DefaultFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
		DefaultFields:   []any{"app", "myapp", "env", "test"},
	})
	handler := NewSlogHandler(logger)
	slogger := slog.New(handler)

	slogger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "app=myapp")
	assert.Contains(t, output, "env=test")
}

func TestSlogHandler_AllLevels(t *testing.T) {
	tests := []struct {
		name          string
		logFunc       func(*slog.Logger)
		expectedLevel string
	}{
		{"debug", func(l *slog.Logger) { l.Debug("msg") }, "level=debug"},
		{"info", func(l *slog.Logger) { l.Info("msg") }, "level=info"},
		{"warn", func(l *slog.Logger) { l.Warn("msg") }, "level=warn"},
		{"error", func(l *slog.Logger) { l.Error("msg") }, "level=error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := New(Opts{
				Writer:          buf,
				Level:           DebugLevel,
				TimestampFormat: time.RFC3339,
			})
			handler := NewSlogHandler(logger)
			slogger := slog.New(handler)

			tt.logFunc(slogger)

			output := buf.String()
			assert.Contains(t, output, tt.expectedLevel)
		})
	}
}

func TestSlogHandler_EmptyGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	// Empty group name should return same handler.
	handlerWithEmptyGroup := handler.WithGroup("")

	assert.Equal(t, handler, handlerWithEmptyGroup)
}

func TestSlogHandler_GroupAttr(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	slogger := slog.New(handler)

	slogger.Info("test message",
		slog.Group("user",
			slog.String("id", "123"),
			slog.String("name", "john"),
		),
	)

	output := buf.String()
	assert.Contains(t, output, "user.id=123")
	assert.Contains(t, output, "user.name=john")
}

func TestSlogHandler_CombinedAttrsAndGroups(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	handlerWithAttrs := handler.WithAttrs([]slog.Attr{slog.String("service", "api")})
	handlerWithGroup := handlerWithAttrs.WithGroup("request")
	slogger := slog.New(handlerWithGroup)

	slogger.Info("test message", "id", "req-123")

	output := buf.String()
	assert.Contains(t, output, "service=api")
	assert.Contains(t, output, "request.id=req-123")
}

func TestSlogLevelToLogf(t *testing.T) {
	tests := []struct {
		slogLevel slog.Level
		expected  Level
	}{
		{slog.LevelDebug, DebugLevel},
		{slog.LevelDebug - 4, DebugLevel}, // Custom debug level.
		{slog.LevelInfo, InfoLevel},
		{slog.LevelWarn, WarnLevel},
		{slog.LevelError, ErrorLevel},
		{slog.LevelError + 4, ErrorLevel}, // Higher than error.
	}

	for _, tt := range tests {
		t.Run(tt.slogLevel.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, slogLevelToLogf(tt.slogLevel))
		})
	}
}

func TestSlogHandler_ConcurrentWrites(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger)
	slogger := slog.New(handler)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				slogger.Info("concurrent log", "goroutine", id, "iteration", j)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1000, len(lines))
}
