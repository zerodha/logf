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

## Why another lib

There are several logging libraries, but the available options didn't meet our use case.

`logf` meets our constraints of:

- Clean API
- Minimal dependencies
- Structured logging but human-readable (`logfmt`!)
- Sane defaults out of the box

## Benchmarks

You can run benchmarks with `make bench`.

### No Colors (Default)

```
BenchmarkNoField-8                       7884771               144.2 ns/op             0 B/op          0 allocs/op
BenchmarkOneField-8                      6251565               186.7 ns/op             0 B/op          0 allocs/op
BenchmarkThreeFields-8                   6273717               188.2 ns/op             0 B/op          0 allocs/op
BenchmarkErrorField-8                    6687260               174.8 ns/op             0 B/op          0 allocs/op
BenchmarkHugePayload-8                   3395139               360.3 ns/op             0 B/op          0 allocs/op
BenchmarkThreeFields_WithCaller-8        2764860               437.9 ns/op           216 B/op          2 allocs/op
```

### With Colors

```
BenchmarkNoField_WithColor-8             6501867               186.6 ns/op             0 B/op          0 allocs/op
BenchmarkOneField_WithColor-8            5938155               205.7 ns/op             0 B/op          0 allocs/op
BenchmarkThreeFields_WithColor-8         4613145               379.4 ns/op             0 B/op          0 allocs/op
BenchmarkErrorField_WithColor-8          3512522               353.6 ns/op             0 B/op          0 allocs/op
BenchmarkHugePayload_WithColor-8         1520659               799.5 ns/op             0 B/op          0 allocs/op
```

For a comparison with existing popular libs, visit [uber-go/zap#performance](https://github.com/uber-go/zap#performance).

## LICENSE

[LICENSE](./LICENSE)
