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

	// Create a logger and add fields which will be logged in every line.
	fields := []any{"request_id", "3MG91VKP", "ip", "1.1.1.1", "method", "method=GET"}
	logger.Info("request success", fields...)
	logger.Warn("this isn't supposed to happen", fields...)

	// Log the error and set exit code as 1.
	logger.Fatal("goodbye world")
}
