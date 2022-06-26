# logf

[![Go Reference](https://pkg.go.dev/badge/github.com/mr-karan/logf.svg)](https://pkg.go.dev/github.com/mr-karan/logf)
[![Go Report Card](https://goreportcard.com/badge/mr-karan/logf)](https://goreportcard.com/report/mr-karan/logf)
[![GitHub Actions](https://github.com/mr-karan/logf/actions/workflows/build.yml/badge.svg)](https://github.com/mr-karan/logf/actions/workflows/build.yml)


logf provides a minimal logging interface for Go applications. It emits **structured logs** ([`logfmt`](https://brandur.org/logfmt) style) in human readable and machine friendly way.

## Example

```go
package main

import (
	"errors"
	"time"

	"github.com/mr-karan/logf"
)

func main() {
	logger := logf.New()
	// Basic log.
	logger.Info("starting app")

	// Change verbosity on the fly.
	logger.SetLevel(logf.DebugLevel)
	logger.Debug("meant for debugging app")

	// Add extra keys to the log.
	logger.WithFields(logf.Fields{
		"component": "api",
		"user":      "karan",
	}).Info("logging with some extra metadata")

	// Log with error key.
	logger.WithError(errors.New("this is a dummy error")).Error("error fetching details")

	// Enable `caller` field in the log and specify the number of frames to skip to get the caller. 
	logger.SetCallerFrame(true, 3)
	// Change the default timestamp format.
	logger.SetTimestampFormat(time.RFC3339Nano)

	// Create a logger and add fields which will be logged in every line.
	requestLogger := logger.WithFields(logf.Fields{"request_id": "3MG91VKP", "ip": "1.1.1.1", "method": "GET"})
	requestLogger.Info("request success")
	requestLogger.Warn("this isn't supposed to happen")

	// Log the error and set exit code as 1.
	logger.Fatal("goodbye world")
}
```

### Text Output

```bash
timestamp=2022-06-26T11:56:46+05:30 level=info message=starting app caller=/home/karan/Code/Personal/logf/examples/main.go:13
timestamp=2022-06-26T11:56:46+05:30 level=debug message=meant for debugging app caller=/home/karan/Code/Personal/logf/examples/main.go:17 level=debug message=meant for debugging app timestamp=2022-06-26T11:56:46+05:30 caller=/home/karan/Code/Personal/logf/examples/main.go:17
timestamp=2022-06-26T11:56:46+05:30 level=info message=logging with some extra metadata component=api user=karan caller=/home/karan/Code/Personal/logf/examples/main.go:23
timestamp=2022-06-26T11:56:46+05:30 level=error message=error fetching details error=this is a dummy error caller=/home/karan/Code/Personal/logf/examples/main.go:26
timestamp=2022-06-26T11:56:46.412189111+05:30 level=info message=request success ip=1.1.1.1 method=GET request_id=3MG91VKP
timestamp=2022-06-26T11:56:46.412204619+05:30 level=warn message=this isn't supposed to happen ip=1.1.1.1 level=warn message=this isn't supposed to happen method=GET request_id=3MG91VKP timestamp=2022-06-26T11:56:46.412204619+05:30
timestamp=2022-06-26T11:56:46.412218628+05:30 level=fatal message=goodbye world ip=1.1.1.1 level=fatal message=goodbye world method=GET request_id=3MG91VKP timestamp=2022-06-26T11:56:46.412218628+05:30
exit status 1
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

```
BenchmarkNoField-8                        290559              3797 ns/op            1576 B/op         74 allocs/op
BenchmarkNoField_NoColor-8               1313766               924.8 ns/op           328 B/op         11 allocs/op
BenchmarkOneField-8                       219285              5445 ns/op            2609 B/op        103 allocs/op
BenchmarkOneField_NoColor-8               668251              1550 ns/op             928 B/op         19 allocs/op
BenchmarkThreeFields-8                    152988              7992 ns/op            3953 B/op        153 allocs/op
BenchmarkThreeFields_NoColor-8            516135              2220 ns/op            1320 B/op         27 allocs/op
BenchmarkHugePayload-8                     57367             22658 ns/op           15121 B/op        356 allocs/op
BenchmarkHugePayload_NoColor-8            140937              7404 ns/op            8342 B/op         62 allocs/op
BenchmarkErrorField-8                     212184              5639 ns/op            2657 B/op        104 allocs/op
BenchmarkErrorField_NoColor-8             703165              1593 ns/op             952 B/op         20 allocs/op
```

For a comparison with existing popular libs, visit [uber-go/zap#performance](https://github.com/uber-go/zap#performance).

## LICENSE

[LICENSE](./LICENSE)
