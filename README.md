<a href="https://zerodha.tech"><img src="https://zerodha.tech/static/images/github-badge.svg" align="right" /></a>

# logf

[![Go Reference](https://pkg.go.dev/badge/github.com/zerodha/logf.svg)](https://pkg.go.dev/github.com/zerodha/logf)
[![Go Report Card](https://goreportcard.com/badge/zerodha/logf)](https://goreportcard.com/report/zerodha/logf)
[![GitHub Actions](https://github.com/zerodha/logf/actions/workflows/build.yml/badge.svg)](https://github.com/zerodha/logf/actions/workflows/build.yml)

logf is a **high performance**, **zero alloc** logging library for Go applications with a **minimal** API overhead. It emits **structured logs** ([`logfmt`](https://brandur.org/logfmt) style) in human readable and machine friendly way. `logf` aims to be customisable without providing an overwhelming amount of things to configure.

## Example

```go
package main

import (
	"os"
	"time"

	"github.com/zerodha/logf"
)

func main() {
	logger := logf.New(os.Stderr)

	// Basic log.
	logger.Info("starting app")

	// Enable colored output.
	logger.SetColorOutput(true)

	// Change verbosity on the fly.
	logger.SetLevel(logf.DebugLevel)
	logger.Debug("meant for debugging app")

	// Add extra keys to the log.
	logger.Info("logging with some extra metadata", "component", "api", "user", "karan")

	// Log with error key.
	logger.Error("error fetching details", "error", "this is a dummy error")

	// Enable `caller` field in the log and specify the number of frames to skip to get the caller.
	logger.SetCallerFrame(true, 3)
	// Change the default timestamp format.
	logger.SetTimestampFormat(time.RFC3339Nano)

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

Agreed there are many logging libraries out there but I was dissatisfied with the current options.

`logf` satisfies my constraints of:

- Clean API
- Minimal Dependencies
- Structured logging but human readable (`logfmt`!)
- Sane defaults out of the box

## Benchmarks

You can run benchmarks with `make bench`.

### No Colors (Default)

```
BenchmarkNoField-8                       7219110               173.0 ns/op             0 B/op          0 allocs/op
BenchmarkOneField-8                      6421900               176.3 ns/op             0 B/op          0 allocs/op
BenchmarkThreeFields-8                   5485582               221.3 ns/op             0 B/op          0 allocs/op
BenchmarkHugePayload-8                    975226              1659 ns/op               0 B/op          0 allocs/op
BenchmarkThreeFields_WithCaller-8        1390599               906.4 ns/op             0 B/op          0 allocs/op
BenchmarkNoField_WithColor-8             1580092               644.2 ns/op             0 B/op          0 allocs/op
```

### With Colors

```
BenchmarkNoField_WithColor-8             1580092               644.2 ns/op             0 B/op          0 allocs/op
BenchmarkOneField_WithColor-8            1810801               689.9 ns/op             0 B/op          0 allocs/op
BenchmarkThreeFields_WithColor-8         1592907               740.8 ns/op             0 B/op          0 allocs/op
BenchmarkHugePayload_WithColor-8          991813              1224 ns/op               0 B/op          0 allocs/op
```

For a comparison with existing popular libs, visit [uber-go/zap#performance](https://github.com/uber-go/zap#performance).

## Contributors

- [Sarat](https://github.com/iamd3vil/)

## LICENSE

[LICENSE](./LICENSE)
