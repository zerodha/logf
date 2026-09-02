package logf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type structuredPayload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type failingJSONMarshaler struct{}

func (failingJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cannot marshal")
}

type nilPanicError struct {
	message string
}

func (e *nilPanicError) Error() string {
	return e.message
}

type panicJSONMarshaler struct{}

func (panicJSONMarshaler) MarshalJSON() ([]byte, error) {
	panic("marshal panic")
}

func TestSlogJSONHandlerConformance(t *testing.T) {
	var output bytes.Buffer
	handler := NewSlogJSONHandler(New(Opts{Writer: &output, Level: InfoLevel}))

	err := slogtest.TestHandler(handler, func() []map[string]any {
		lines := strings.Split(strings.TrimSpace(output.String()), "\n")
		results := make([]map[string]any, 0, len(lines))
		for _, line := range lines {
			var result map[string]any
			require.NoError(t, json.Unmarshal([]byte(line), &result))
			results = append(results, result)
		}
		return results
	})
	require.NoError(t, err)
}

func TestSlogJSONHandlerValues(t *testing.T) {
	var output bytes.Buffer
	logger := New(Opts{
		Writer:        &output,
		Level:         DebugLevel,
		EnableColor:   true,
		DefaultFields: []any{"app", "logf"},
	})
	handler := NewSlogJSONHandler(logger)
	logger.DefaultFields[1] = "changed"
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 123_456_789, time.UTC)
	var nilError *nilPanicError
	record := slog.NewRecord(recordTime, slog.LevelInfo, "request <accepted>", 0)
	record.AddAttrs(
		slog.String("text", "<&\n"),
		slog.Int64("int", -42),
		slog.Uint64("uint", ^uint64(0)),
		slog.Float64("float", 1e-7),
		slog.Bool("bool", true),
		slog.Duration("duration", 5*time.Second),
		slog.Time("created_at", recordTime),
		slog.Any("error", io.EOF),
		slog.Any("bytes", []byte("logf")),
		slog.Any("payload", structuredPayload{ID: 7, Name: "worker"}),
		slog.Float64("nan", math.NaN()),
		slog.Any("unsupported", make(chan int)),
		slog.Any("marshal_error", failingJSONMarshaler{}),
		slog.Any("nil_error", nilError),
		slog.Any("marshal_panic", panicJSONMarshaler{}),
	)

	require.NoError(t, handler.Handle(context.Background(), record))
	line := strings.TrimSpace(output.String())
	assert.Contains(t, line, `"msg":"request <accepted>"`)
	assert.Contains(t, line, `"text":"<&\n"`)

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(line), &fields))
	assert.JSONEq(t, `"2024-01-15T10:30:00.123456789Z"`, string(fields[slog.TimeKey]))
	assert.JSONEq(t, `"INFO"`, string(fields[slog.LevelKey]))
	assert.JSONEq(t, `"request <accepted>"`, string(fields[slog.MessageKey]))
	assert.JSONEq(t, `"logf"`, string(fields["app"]))
	assert.Equal(t, `-42`, string(fields["int"]))
	assert.Equal(t, `18446744073709551615`, string(fields["uint"]))
	assert.Equal(t, `1e-7`, string(fields["float"]))
	assert.Equal(t, `true`, string(fields["bool"]))
	assert.Equal(t, `5000000000`, string(fields["duration"]))
	assert.JSONEq(t, `"2024-01-15T10:30:00.123456789Z"`, string(fields["created_at"]))
	assert.JSONEq(t, `"EOF"`, string(fields["error"]))
	assert.JSONEq(t, `"bG9nZg=="`, string(fields["bytes"]))
	assert.JSONEq(t, `{"id":7,"name":"worker"}`, string(fields["payload"]))
	assert.JSONEq(t, `"!ERROR:json: unsupported value: NaN"`, string(fields["nan"]))
	assert.Contains(t, string(fields["unsupported"]), `!ERROR:json: unsupported type: chan int`)
	assert.Contains(t, string(fields["marshal_error"]), `!ERROR:json: error calling MarshalJSON`)
	assert.JSONEq(t, `"<nil>"`, string(fields["nil_error"]))
	assert.JSONEq(t, `"!PANIC: marshal panic"`, string(fields["marshal_panic"]))
}

func TestSlogJSONHandlerPersistentGroups(t *testing.T) {
	var output bytes.Buffer
	handler := NewSlogJSONHandler(New(Opts{Writer: &output, Level: DebugLevel})).WithAttrs([]slog.Attr{
		slog.String("root", "yes"),
	}).WithGroup("request").WithAttrs([]slog.Attr{
		slog.String("service", "api"),
		slog.Group("client", slog.String("name", "web")),
	}).WithGroup("nested")
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	record := slog.NewRecord(recordTime, slog.LevelInfo, "first", 0)
	record.AddAttrs(
		slog.Group("", slog.String("id", "req-1")),
		slog.Group("empty"),
	)
	require.NoError(t, handler.Handle(context.Background(), record))
	assert.Equal(t,
		`{"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"first","root":"yes","request":{"service":"api","client":{"name":"web"},"nested":{"id":"req-1"}}}`+"\n",
		output.String(),
	)

	output.Reset()
	record = slog.NewRecord(recordTime, slog.LevelInfo, "second", 0)
	require.NoError(t, handler.Handle(context.Background(), record))
	assert.Equal(t,
		`{"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"second","root":"yes","request":{"service":"api","client":{"name":"web"}}}`+"\n",
		output.String(),
	)
}

func TestSlogJSONHandlerLogValuerTiming(t *testing.T) {
	var output bytes.Buffer
	persistentCalls := 0
	persistent := NewSlogJSONHandler(New(Opts{Writer: &output, Level: DebugLevel})).WithAttrs([]slog.Attr{
		slog.Any("persistent", countingLogValuer{calls: &persistentCalls}),
	})
	assert.Equal(t, 1, persistentCalls)

	dynamicCalls := 0
	logger := slog.New(persistent)
	logger.Info("first", slog.Any("dynamic", countingLogValuer{calls: &dynamicCalls}))
	logger.Info("second", slog.Any("dynamic", countingLogValuer{calls: &dynamicCalls}))
	assert.Equal(t, 1, persistentCalls)
	assert.Equal(t, 2, dynamicCalls)

	disabledCalls := 0
	disabled := slog.New(NewSlogJSONHandler(New(Opts{Writer: io.Discard, Level: ErrorLevel})))
	disabled.Info("disabled", slog.Any("value", countingLogValuer{calls: &disabledCalls}))
	assert.Zero(t, disabledCalls)
}

func TestSlogJSONHandlerSourceZeroTimeAndWriterError(t *testing.T) {
	var output bytes.Buffer
	handler := NewSlogJSONHandler(New(Opts{
		Writer:       &output,
		Level:        DebugLevel,
		EnableCaller: true,
	}))
	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])
	frame, _ := runtime.CallersFrames(pcs[:]).Next()
	record := slog.NewRecord(time.Time{}, slog.LevelWarn, "test", pcs[0])

	require.NoError(t, handler.Handle(context.Background(), record))
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(output.Bytes()), &fields))
	assert.NotContains(t, fields, slog.TimeKey)
	assert.JSONEq(t, `"WARN"`, string(fields[slog.LevelKey]))
	assert.JSONEq(t,
		`{"function":`+strconv.Quote(frame.Function)+`,"file":`+strconv.Quote(frame.File)+`,"line":`+strconv.Itoa(frame.Line)+`}`,
		string(fields[slog.SourceKey]),
	)

	errorHandler := NewSlogJSONHandler(New(Opts{Writer: &errWriter{}, Level: DebugLevel}))
	err := errorHandler.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
	require.EqualError(t, err, "dummy error")
}

func TestSlogJSONHandlerConcurrentWrites(t *testing.T) {
	var output bytes.Buffer
	handler := NewSlogJSONHandler(New(Opts{Writer: &output, Level: DebugLevel})).WithAttrs([]slog.Attr{
		slog.String("service", "api"),
	})
	logger := slog.New(handler)

	var wait sync.WaitGroup
	for id := range 10 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := range 100 {
				logger.Info("event", "worker", id, "iteration", iteration)
			}
		}()
	}
	wait.Wait()

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.Len(t, lines, 1000)
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		var fields map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &fields))
		assert.Equal(t, "api", fields["service"])
		pair := strconv.Itoa(int(fields["worker"].(float64))) + "/" + strconv.Itoa(int(fields["iteration"].(float64)))
		_, duplicate := seen[pair]
		assert.False(t, duplicate, "duplicate record %s", pair)
		seen[pair] = struct{}{}
	}
	assert.Len(t, seen, 1000)
}
