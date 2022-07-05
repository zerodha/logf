package logf

import "os"

type FieldLogger struct {
	fields []F
	logger *Logger
}

func (l *FieldLogger) Debug(msg string) {
	l.logger.handleLog(msg, DebugLevel, l.fields...)
}

// Info emits a info log line.
func (l *FieldLogger) Info(msg string) {
	l.logger.handleLog(msg, InfoLevel, l.fields...)
}

// Warn emits a warning log line.
func (l *FieldLogger) Warn(msg string) {
	l.logger.handleLog(msg, WarnLevel, l.fields...)
}

// Error emits an error log line.
func (l *FieldLogger) Error(msg string) {
	l.logger.handleLog(msg, ErrorLevel, l.fields...)
}

// Fatal emits a fatal level log line.
// It aborts the current program with an exit code of 1.
func (l *FieldLogger) Fatal(msg string) {
	l.logger.handleLog(msg, ErrorLevel, l.fields...)
	os.Exit(1)
}
