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

	// Enable colored output.
	logger.SetColorOutput(true)

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
	requestLogger := logger.WithFields(logf.Fields{"request_id": "3MG91VKP", "ip": "1.1.1.1", "method": "method=GET"})
	requestLogger.Info("request success")
	requestLogger.Warn("this isn't supposed to happen")

	// Log the error and set exit code as 1.
	logger.Fatal("goodbye world")
}
