package logf

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelParsing(t *testing.T) {
	cases := []struct {
		String string
		Lvl    Level
		Num    int
	}{
		{"debug", DebugLevel, 0},
		{"info", InfoLevel, 1},
		{"warn", WarnLevel, 2},
		{"error", ErrorLevel, 3},
		{"fatal", FatalLevel, 4},
	}

	for _, c := range cases {
		t.Run(c.String, func(t *testing.T) {
			assert.Equal(t, c.Lvl.String(), c.String, "level should be equal")
		})
	}

	// Check for an invalid case.
	t.Run("invalid", func(t *testing.T) {
		var invalidLvl Level = 10
		assert.Equal(t, invalidLvl.String(), "invalid lvl", "invalid level")
	})
}

func TestNewLoggerDefault(t *testing.T) {
	l := New(os.Stderr)
	assert.Equal(t, l.level, InfoLevel, "level is info")
	assert.Equal(t, l.enableColor, false, "color output is disabled")
	assert.Equal(t, l.enableCaller, false, "caller is disabled")
	assert.Equal(t, l.callerSkipFrameCount, 0, "skip frame count is 0")
	assert.Equal(t, l.tsFormat, defaultTSFormat, "timestamp format is default")
}

func TestLogFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(buf)
	l.SetColorOutput(false)

	// Info log.
	l.Info("hello world")
	assert.Contains(t, buf.String(), `level=info message="hello world"`, "info log")
	buf.Reset()

	// Log with field.
	l.Warn("testing fields", "stack", "testing")
	assert.Contains(t, buf.String(), `level=warn message="testing fields" stack=testing`, "warning log")
	buf.Reset()

	// Log with error.
	fakeErr := errors.New("this is a fake error")
	l.Error("testing error", "error", fakeErr)
	assert.Contains(t, buf.String(), `level=error message="testing error" error="this is a fake error"`, "error log")
	buf.Reset()
}

func TestOddNumberedFields(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(buf)
	l.SetColorOutput(false)

	// Give a odd number of fields.
	l.Info("hello world", "key1", "val1", "key2")
	assert.Contains(t, buf.String(), `level=info message="hello world" key1=val1`)
	buf.Reset()
}

// These test are typically meant to be run with the data race detector.
func TestLoggerConcurrency(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(buf)
	l.SetColorOutput(false)

	for _, n := range []int{10, 100, 1000} {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() { genLogs(l); wg.Done() }()
		}
		wg.Wait()
	}
}

func genLogs(l Logger) {
	for i := 0; i < 100; i++ {
		l.Info("random log", "index", strconv.FormatInt(int64(i), 10))
	}
}
