<a href="https://zerodha.tech"><img src="https://zerodha.tech/static/images/github-badge.svg" align="right" /></a>

# 💥 logf

[![Go Reference](https://pkg.go.dev/badge/github.com/zerodha/logf.svg)](https://pkg.go.dev/github.com/zerodha/logf)
[![Go Report Card](https://goreportcard.com/badge/zerodha/logf)](https://goreportcard.com/report/zerodha/logf)
[![GitHub Actions](https://github.com/zerodha/logf/actions/workflows/build.yml/badge.svg)](https://github.com/zerodha/logf/actions/workflows/build.yml)

`logf` is a high-performance, zero-alloc logging library for Go applications with a minimal API overhead. It's also the fastest logfmt logging library for Go.

`logf` emits structured logs in [`logfmt`](https://brandur.org/logfmt) style. `logfmt` is a flexible format that involves `key=value` pairs to emit structured log lines. `logfmt` achieves the goal of generating logs that are not just machine-friendly but also readable by humans, unlike the clunky JSON lines.

## Example

```go
package main

import (
	"time"

	"github.com/zerodha/logf"
)

func main() {
	logger := logf.New(logf.Opts{
		EnableColor:          true,
		Level:                logf.DebugLevel,
		CallerSkipFrameCount: 3,
		EnableCaller:         true,
		TimestampFormat:      time.RFC3339Nano,
		DefaultFields:        []any{"scope", "example"},
	})

	// Basic logs.
	logger.Info("starting app")
	logger.Debug("meant for debugging app")

	// Add extra keys to the log.
	logger.Info("logging with some extra metadata", "component", "api", "user", "karan")

	// Log with error key.
	logger.Error("error fetching details", "error", "this is a dummy error")

	// Log the error and set exit code as 1.
	logger.Fatal("goodbye world")
}
```

### Text Output

```bash
timestamp=2022-07-07T12:09:10.221+05:30 level=info message="starting app"
timestamp=2022-07-07T12:09:10.221+05:30 level=info message="logging with some extra metadata" component=api user=karan
timestamp=2022-07-07T12:09:10.221+05:30 level=error message="error fetching details" error="this is a dummy error"
timestamp=2022-07-07T12:09:10.221+05:30 level=fatal message="goodbye world"
```

### Console Output

![](examples/screenshot.png)

## Slog Adapter

`logf` provides a `slog.Handler` implementation, allowing you to use `logf` as a backend for Go's standard `log/slog` package. This gives you the best of both worlds: the familiar `slog` API with `logf`'s efficient logfmt output.

```go
package main

import (
	"errors"
	"log/slog"
	"time"

	"github.com/zerodha/logf"
)

func main() {
	// Create a logf logger.
	logger := logf.New(logf.Opts{
		EnableColor:     true,
		Level:           logf.DebugLevel,
		TimestampFormat: time.RFC3339Nano,
	})

	// Create a slog handler using logf.
	handler := logf.NewSlogHandler(logger)

	// Create a slog.Logger.
	slogger := slog.New(handler)

	// Use standard slog API.
	slogger.Info("user logged in", "user_id", 12345, "username", "johndoe")

	// Typed attributes.
	slogger.Info("request completed",
		slog.String("method", "GET"),
		slog.Int("status", 200),
		slog.Duration("latency", 150*time.Millisecond),
	)

	// Groups.
	slogger.Info("http request",
		slog.Group("request",
			slog.String("method", "POST"),
			slog.String("path", "/api/users"),
		),
	)

	// WithAttrs for persistent fields.
	requestLogger := slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("request_id", "abc-123"),
	}))
	requestLogger.Info("processing request")

	// WithGroup for namespaced logging.
	dbLogger := slog.New(handler.WithGroup("database"))
	dbLogger.Info("query executed", "rows", 42)

	// Error logging.
	err := errors.New("connection timeout")
	slogger.Error("database error", "error", err)
}
```

**Output:**
```
timestamp=2024-01-15T10:30:00.000Z level=info message="user logged in" user_id=12345 username=johndoe
timestamp=2024-01-15T10:30:00.000Z level=info message="request completed" method=GET status=200 latency=150ms
timestamp=2024-01-15T10:30:00.000Z level=info message="http request" request.method=POST request.path=/api/users
timestamp=2024-01-15T10:30:00.000Z level=info message="processing request" request_id=abc-123
timestamp=2024-01-15T10:30:00.000Z level=info message="query executed" database.rows=42
timestamp=2024-01-15T10:30:00.000Z level=error message="database error" error="connection timeout"
```

### JSON output

Use `NewSlogJSONHandler` for line-delimited structured JSON:

```go
handler := logf.NewSlogJSONHandler(logger)
slogger := slog.New(handler)

slogger.Info("request completed",
	slog.String("method", "GET"),
	slog.Int("status", 200),
	slog.Group("response", slog.Duration("latency", 150*time.Millisecond)),
)
```

```json
{"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"request completed","method":"GET","status":200,"response":{"latency":150000000}}
```

The JSON handler follows `slog.JSONHandler` conventions: the built-in keys are `time`, `level`, `msg`, and optional `source`; groups are nested objects; durations are nanoseconds; arbitrary values retain their `encoding/json` structure; and HTML characters are not escaped. `EnableColor` and `TimestampFormat` do not affect JSON output.

Following slog's [performance guidance](https://pkg.go.dev/log/slog#hdr-Performance_considerations), attributes added with `Logger.With` or `Handler.WithAttrs` are resolved and formatted once. Per-record `LogValuer` values are resolved only for enabled records. Each complete record is passed to the synchronized writer in one call.

## Why another lib

There are several logging libraries, but the available options didn't meet our use case.

`logf` meets our constraints of:

- Clean API
- Minimal dependencies
- Structured logging but human-readable (`logfmt`!)
- Sane defaults out of the box

## Benchmarks

You can run benchmarks with `make bench`.

### Logf Direct (zero allocations)

```
BenchmarkNoField-20                       1889884               681.5 ns/op             0 B/op          0 allocs/op
BenchmarkOneField-20                      1700284               762.1 ns/op             0 B/op          0 allocs/op
BenchmarkThreeFields-20                   1533897               789.6 ns/op             0 B/op          0 allocs/op
BenchmarkErrorField-20                    1725673               686.0 ns/op             0 B/op          0 allocs/op
BenchmarkHugePayload-20                   1250776               923.0 ns/op             0 B/op          0 allocs/op
BenchmarkThreeFields_WithCaller-20        1244536              1049   ns/op           248 B/op          2 allocs/op
```

### Slog Text Handler Comparison

Representative `go test -run '^$' -bench '^(BenchmarkSlog|BenchmarkVanillaSlog_Text)' -benchmem -count=5` medians on Go 1.27:

| Benchmark | logf slog | slog Text |
|-----------|-----------|-----------|
| NoField | 112 ns/op, 0 allocs | 149 ns/op, 0 allocs |
| ThreeFields | 147 ns/op, 0 allocs | 164 ns/op, 0 allocs |
| HugePayload | 294 ns/op, 1 alloc | 509 ns/op, 1 alloc |
| WithAttrs | 130 ns/op, 0 allocs | 167 ns/op, 0 allocs |
| WithAttrsNestedGroup | 125 ns/op, 0 allocs | 147 ns/op, 0 allocs |
| Disabled | **0.85 ns/op, 0 allocs** | - |

`logf` preformats persistent attributes and short-circuits disabled records before they allocate a slog Record. Results vary by Go version, CPU, and output writer.

### Slog JSON Handler Comparison

Representative `go test -run '^$' -bench '^(BenchmarkSlogJSON|BenchmarkVanillaSlog_JSON)' -benchmem -count=5` medians on Go 1.27:

| Benchmark | logf JSON | slog JSON |
|-----------|------------|-----------|
| NoField | 142 ns/op, 0 allocs | 180 ns/op, 0 allocs |
| ThreeFields | 157 ns/op, 0 allocs | 197 ns/op, 0 allocs |
| HugePayload | 350 ns/op, 1 alloc | 509 ns/op, 5 allocs |
| WithAttrs | 161 ns/op, 0 allocs | 166 ns/op, 0 allocs |
| NestedGroup | 248 ns/op, 5 allocs | 293 ns/op, 5 allocs |
| StructuredAny | 252 ns/op, 2 allocs | 291 ns/op, 2 allocs |
| Disabled LogValuer | 1.18 ns/op, 0 allocs | 1.39 ns/op, 0 allocs |

The handler-only caller path is 231 ns/op with zero allocations, compared with 738 ns/op, 584 B, and six allocations for `slog.JSONHandler`.

For a comparison with existing popular libs, visit [uber-go/zap#performance](https://github.com/uber-go/zap#performance).

## LICENSE

[LICENSE](./LICENSE)
