package logf

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type countingLogValuer struct {
	calls *int
}

func (v countingLogValuer) LogValue() slog.Value {
	*v.calls++
	return slog.StringValue("resolved")
}

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
		name        string
		loggerLevel Level
		slogLevel   slog.Level
		wantEnabled bool
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
			assert.Equal(t, tt.wantEnabled, handler.Enabled(context.Background(), tt.slogLevel))
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
	logger.DefaultFields[1] = "changed-after-handler-construction"
	record := slog.NewRecord(
		time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		slog.LevelInfo,
		"test",
		0,
	)

	assert.NoError(t, handler.Handle(context.Background(), record))
	assert.Equal(t,
		"timestamp=2024-01-15T10:30:00Z level=info message=test app=myapp env=test\n",
		buf.String(),
	)
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

func TestSlogHandler_PersistentAttrsKeepGroupOrder(t *testing.T) {
	const timestamp = "2024-01-15T10:30:00Z"
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("attrs before group", func(t *testing.T) {
		buf := &bytes.Buffer{}
		handler := NewSlogHandler(New(Opts{
			Writer:          buf,
			Level:           DebugLevel,
			TimestampFormat: time.RFC3339,
		})).WithAttrs([]slog.Attr{
			slog.String("service", "api"),
			slog.Group("client", slog.String("name", "web")),
		}).WithGroup("request")
		record := slog.NewRecord(recordTime, slog.LevelInfo, "test", 0)
		record.AddAttrs(slog.Group("http", slog.Group("response", slog.Int("status", 200))))

		assert.NoError(t, handler.Handle(context.Background(), record))
		assert.Equal(t,
			"timestamp="+timestamp+" level=info message=test service=api client.name=web request.http.response.status=200\n",
			buf.String(),
		)
	})

	t.Run("group before attrs", func(t *testing.T) {
		buf := &bytes.Buffer{}
		handler := NewSlogHandler(New(Opts{
			Writer:          buf,
			Level:           DebugLevel,
			TimestampFormat: time.RFC3339,
		})).WithGroup("request").WithAttrs([]slog.Attr{
			slog.String("service", "api"),
			slog.Group("client", slog.String("name", "web")),
		})
		record := slog.NewRecord(recordTime, slog.LevelInfo, "test", 0)

		assert.NoError(t, handler.Handle(context.Background(), record))
		assert.Equal(t,
			"timestamp="+timestamp+" level=info message=test request.service=api request.client.name=web\n",
			buf.String(),
		)
	})
}

func TestSlogHandler_InlineEmptyAndQuotedGroups(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewSlogHandler(New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	}))
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	record := slog.NewRecord(recordTime, slog.LevelInfo, "test", 0)
	record.AddAttrs(
		slog.Group("outer",
			slog.Group("", slog.String("inline", "ok")),
			slog.Group("empty"),
		),
		slog.Group("", slog.String("top", "yes")),
		slog.Group("http request", slog.Int("status=code", 200)),
	)

	assert.NoError(t, handler.Handle(context.Background(), record))
	assert.Equal(t,
		"timestamp=2024-01-15T10:30:00Z level=info message=test outer.inline=ok top=yes \"http request.status=code\"=200\n",
		buf.String(),
	)
}

func TestSlogHandler_WithAttrsResolveOnce(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		TimestampFormat: time.RFC3339,
	})
	directCalls := 0
	nestedCalls := 0
	handler := NewSlogHandler(logger).WithAttrs([]slog.Attr{
		slog.Any("value", countingLogValuer{calls: &directCalls}),
		slog.Group("request", slog.Any("value", countingLogValuer{calls: &nestedCalls})),
	})

	assert.Equal(t, 1, directCalls)
	assert.Equal(t, 1, nestedCalls)

	slogger := slog.New(handler)
	slogger.Info("first")
	slogger.Info("second")

	assert.Equal(t, 1, directCalls)
	assert.Equal(t, 1, nestedCalls)
	assert.Equal(t, 2, strings.Count(buf.String(), " value=resolved "))
	assert.Equal(t, 2, strings.Count(buf.String(), "request.value=resolved"))
}

func TestSlogHandler_ColoredPersistentAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewSlogHandler(New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		EnableColor:     true,
		TimestampFormat: time.RFC3339,
	})).WithAttrs([]slog.Attr{
		slog.String("service", "api"),
		slog.Group("request", slog.String("id", "req-123")),
	})
	record := slog.NewRecord(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), slog.LevelInfo, "test", 0)

	assert.NoError(t, handler.Handle(context.Background(), record))
	assert.Equal(t,
		cyan+"timestamp="+reset+"2024-01-15T10:30:00Z "+
			cyan+"level"+reset+"=info "+
			cyan+"message"+reset+"=test "+
			cyan+"service"+reset+"=api "+
			cyan+"request.id"+reset+"=req-123\n",
		buf.String(),
	)
}

func TestSlogHandler_RecordTimeUint64AndCaller(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewSlogHandler(New(Opts{
		Writer:          buf,
		Level:           DebugLevel,
		EnableCaller:    true,
		TimestampFormat: time.RFC3339,
	}))
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])
	frame, _ := runtime.CallersFrames(pcs[:]).Next()
	record := slog.NewRecord(recordTime, slog.LevelInfo, "test", pcs[0])
	record.AddAttrs(slog.Uint64("id", ^uint64(0)))

	assert.NoError(t, handler.Handle(context.Background(), record))
	assert.Equal(t,
		"timestamp=2024-01-15T10:30:00Z level=info message=test caller="+
			frame.File+":"+strconv.Itoa(frame.Line)+" id=18446744073709551615\n",
		buf.String(),
	)

	buf.Reset()
	record = slog.NewRecord(recordTime, slog.LevelInfo, "test", 0)
	assert.NoError(t, handler.Handle(context.Background(), record))
	assert.Equal(t, "timestamp=2024-01-15T10:30:00Z level=info message=test\n", buf.String())
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
		EnableCaller:    true,
		TimestampFormat: time.RFC3339,
	})
	handler := NewSlogHandler(logger).WithAttrs([]slog.Attr{slog.String("service", "api")})
	slogger := slog.New(handler)

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range 100 {
				slogger.Info("concurrent log", "goroutine", id, "iteration", j)
			}
		}(i)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 1000)
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		assert.Equal(t, 1, strings.Count(line, "service=api"))

		var goroutine, iteration string
		for _, field := range strings.Fields(line) {
			switch {
			case strings.HasPrefix(field, "goroutine="):
				goroutine = field
			case strings.HasPrefix(field, "iteration="):
				iteration = field
			}
		}
		assert.Equal(t, 1, strings.Count(line, "caller="))
		assert.NotEmpty(t, goroutine)
		assert.NotEmpty(t, iteration)

		pair := goroutine + "/" + iteration
		_, duplicate := seen[pair]
		assert.False(t, duplicate, "duplicate record %s", pair)
		seen[pair] = struct{}{}
	}
	assert.Len(t, seen, 1000)
}
