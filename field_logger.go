package logf

type FieldLogger struct {
	fields Fields
	logger *Logger
}

func (l *FieldLogger) Debug(msg string) {
	l.logger.handleLog(msg, DebugLevel, l.fields)
}

// Info emits a info log line.
func (l *FieldLogger) Info(msg string) {
	l.logger.handleLog(msg, InfoLevel, l.fields)
}

// Warn emits a warning log line.
func (l *FieldLogger) Warn(msg string) {
	l.logger.handleLog(msg, WarnLevel, l.fields)
}

// Error emits an error log line.
func (l *FieldLogger) Error(msg string) {
	l.logger.handleLog(msg, ErrorLevel, l.fields)
}
