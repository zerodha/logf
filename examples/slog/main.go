package main

import (
	"errors"
	"log/slog"
	"time"

	"github.com/zerodha/logf"
)

func main() {
	// Create a logf logger with desired options.
	logger := logf.New(logf.Opts{
		EnableColor:     true,
		Level:           logf.DebugLevel,
		EnableCaller:    true,
		TimestampFormat: time.RFC3339Nano,
		DefaultFields:   []any{"app", "myapp"},
	})

	// Create a slog handler using the logf logger.
	handler := logf.NewSlogHandler(logger)

	// Create a slog.Logger using the handler.
	slogger := slog.New(handler)

	// Basic logging.
	slogger.Info("starting application")
	slogger.Debug("debug information")

	// Logging with additional attributes.
	slogger.Info("user logged in",
		"user_id", 12345,
		"username", "johndoe",
	)

	// Using typed attributes.
	slogger.Info("request completed",
		slog.String("method", "GET"),
		slog.Int("status", 200),
		slog.Duration("latency", 150*time.Millisecond),
	)

	// Error logging.
	err := errors.New("connection timeout")
	slogger.Error("database error",
		"error", err,
		"retry_count", 3,
	)

	// Using groups for structured data.
	slogger.Info("http request",
		slog.Group("request",
			slog.String("method", "POST"),
			slog.String("path", "/api/users"),
		),
		slog.Group("response",
			slog.Int("status", 201),
			slog.Duration("duration", 45*time.Millisecond),
		),
	)

	// Using WithAttrs to add persistent attributes.
	requestLogger := slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("request_id", "abc-123"),
		slog.String("trace_id", "xyz-789"),
	}))
	requestLogger.Info("processing request")
	requestLogger.Info("request completed")

	// Using WithGroup for namespaced logging.
	dbLogger := slog.New(handler.WithGroup("database"))
	dbLogger.Info("query executed",
		"query", "SELECT * FROM users",
		"rows", 42,
	)

	// Chaining WithGroup for nested namespaces.
	httpDbLogger := slog.New(handler.WithGroup("http").WithGroup("database"))
	httpDbLogger.Warn("slow query detected",
		"duration_ms", 1500,
	)

	// Different log levels.
	slogger.Debug("this is a debug message")
	slogger.Info("this is an info message")
	slogger.Warn("this is a warning message")
	slogger.Error("this is an error message")
}
