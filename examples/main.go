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
		DefaultFields:        []interface{}{"scope", "example"},
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
