package logf

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogFormatWithEnableCaller(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf, EnableCaller: true})

	l.Info("hello world")
	require.Contains(t, buf.String(), `level=info message="hello world" caller=`)
	require.Contains(t, buf.String(), `logf/log_test.go:19`)
	buf.Reset()

	lC := New(Opts{Writer: buf, EnableCaller: true, EnableColor: true})
	lC.Info("hello world")
	require.Contains(t, buf.String(), `logf/log_test.go:25`)
	buf.Reset()
}

func TestLevelParsing(t *testing.T) {
	cases := []struct {
		String string
		Lvl    Level
		Num    int
	}{
		{"debug", DebugLevel, 1},
		{"info", InfoLevel, 2},
		{"warn", WarnLevel, 3},
		{"error", ErrorLevel, 4},
		{"fatal", FatalLevel, 5},
	}

	for _, c := range cases {
		t.Run(c.String, func(t *testing.T) {
			require.Equal(t, c.Lvl.String(), c.String, "level should be equal")
		})
	}

	// Test LevelFromString.
	for _, c := range cases {
		t.Run(fmt.Sprintf("from-string-%v", c.String), func(t *testing.T) {
			str, err := LevelFromString(c.String)
			if err != nil {
				t.Fatalf("error parsing level: %v", err)
			}
			require.Equal(t, c.Lvl, str, "level should be equal")
		})
	}

	// Check for an invalid case.
	t.Run("invalid", func(t *testing.T) {
		var invalidLvl Level = 10
		require.Equal(t, invalidLvl.String(), "invalid lvl", "invalid level")
	})
}

func TestNewLoggerDefault(t *testing.T) {
	l := New(Opts{})
	require.Equal(t, l.Opts.Level, InfoLevel, "level is info")
	require.Equal(t, l.Opts.EnableColor, false, "color output is disabled")
	require.Equal(t, l.Opts.EnableCaller, false, "caller is disabled")
	require.Equal(t, l.Opts.CallerSkipFrameCount, 3, "skip frame count is 3")
	require.Equal(t, l.Opts.TimestampFormat, defaultTSFormat, "timestamp format is default")
}

func TestNewSyncWriterWithNil(t *testing.T) {
	w := newSyncWriter(nil)
	require.NotNil(t, w.w, "writer should not be nil")
}

func TestLogFormat(t *testing.T) {
	buf := &bytes.Buffer{}

	l := New(Opts{Writer: buf, Level: DebugLevel})
	// Debug log.
	l.Debug("debug log")
	require.Contains(t, buf.String(), `level=debug message="debug log"`)
	buf.Reset()

	l = New(Opts{Writer: buf})

	// Debug log but with default level set to info.
	l.Debug("debug log")
	require.NotContains(t, buf.String(), `level=debug message="debug log"`)
	buf.Reset()

	// Info log.
	l.Info("hello world")
	require.Contains(t, buf.String(), `level=info message="hello world"`, "info log")
	buf.Reset()

	// Log with field.
	l.Warn("testing fields", "stack", "testing")
	require.Contains(t, buf.String(), `level=warn message="testing fields" stack=testing`, "warning log")
	buf.Reset()

	// Log with error.
	fakeErr := errors.New("this is a fake error")
	l.Error("testing error", "error", fakeErr)
	require.Contains(t, buf.String(), `level=error message="testing error" error="this is a fake error"`, "error log")
	buf.Reset()

	// Fatal log
	var hadExit bool
	exit = func() {
		hadExit = true
	}

	l.Fatal("fatal log")
	require.True(t, hadExit, "exit should have been called")
	require.Contains(t, buf.String(), `level=fatal message="fatal log"`, "fatal log")
	buf.Reset()
}

func TestLogFormatWithColor(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf, EnableColor: true})

	// Info log.
	l.Info("hello world")
	require.Contains(t, buf.String(), "\x1b[36mlevel\x1b[0m=info \x1b[36mmessage\x1b[0m=\"hello world\" \n")
	buf.Reset()
}

func TestLoggerTypes(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf, Level: DebugLevel})
	type foo struct {
		A int
	}
	l.Info("hello world",
		"string", "foo",
		"int", 1,
		"int8", int8(1),
		"int16", int16(1),
		"int32", int32(1),
		"int64", int64(1),
		"float32", float32(1.0),
		"float64", float64(1.0),
		"struct", foo{A: 1},
		"bool", true,
	)

	require.Contains(t, buf.String(), "level=info message=\"hello world\" string=foo int=1 int8=1 int16=1 int32=1 int64=1 float32=1 float64=1 struct={1} bool=true \n")
}

func TestLogFormatWithDefaultFields(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf, DefaultFields: []interface{}{"defaultkey", "defaultvalue"}})

	l.Info("hello world")
	require.Contains(t, buf.String(), `level=info message="hello world" defaultkey=defaultvalue`)
	buf.Reset()

	l.Info("hello world", "component", "logf")
	require.Contains(t, buf.String(), `level=info message="hello world" defaultkey=defaultvalue component=logf`)
	buf.Reset()
}

type errWriter struct{}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("dummy error")
}

func TestIoWriterError(t *testing.T) {
	w := &errWriter{}
	l := New(Opts{Writer: w})
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(0)
	l.Info("hello world")
	require.Contains(t, buf.String(), "error logging: dummy error\n")
}

func TestWriteQuotedStringCases(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf})

	// cases from
	// https://github.com/go-logfmt/logfmt/blob/99455b83edb21b32a1f1c0a32f5001b77487b721/encode_test.go
	data := []struct {
		key, value interface{}
		want       string
	}{
		{key: "k", value: "v", want: "k=v"},
		{key: "k", value: nil, want: "k=null"},
		{key: `\`, value: "v", want: `\=v`},
		{key: "k", value: "", want: "k="},
		{key: "k", value: "null", want: `k="null"`},
		{key: "k", value: "<nil>", want: `k=<nil>`},
		{key: "k", value: true, want: "k=true"},
		{key: "k", value: 1, want: "k=1"},
		{key: "k", value: 1.025, want: "k=1.025"},
		{key: "k", value: 1e-3, want: "k=0.001"},
		{key: "k", value: 3.5 + 2i, want: "k=(3.5+2i)"},
		{key: "k", value: "v v", want: `k="v v"`},
		{key: "k", value: " ", want: `k=" "`},
		{key: "k", value: `"`, want: `k="\""`},
		{key: "k", value: `=`, want: `k="="`},
		{key: "k", value: `\`, want: `k=\`},
		{key: "k", value: `=\`, want: `k="=\\"`},
		{key: "k", value: `\"`, want: `k="\\\""`},
		{key: "k", value: "\xbd", want: `k="\ufffd"`},
		{key: "k", value: "\ufffd\x00", want: `k="\ufffd\u0000"`},
		{key: "k", value: "\ufffd", want: `k="\ufffd"`},
		{key: "k", value: []byte("\ufffd\x00"), want: `k="\ufffd\u0000"`},
		{key: "k", value: []byte("\ufffd"), want: `k="\ufffd"`},
	}

	for _, d := range data {
		l.Info("hello world", d.key, d.value)
		require.Contains(t, buf.String(), d.want)
		buf.Reset()
	}
}

func TestOddNumberedFields(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf})

	// Give a odd number of fields.
	l.Info("hello world", "key1", "val1", "key2")
	require.Contains(t, buf.String(), `level=info message="hello world" key1=val1`)
	buf.Reset()
}

func TestOddNumberedFieldsWithDefaultFields(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf, DefaultFields: []interface{}{
		"defaultkey", "defaultval",
	}})

	// Give a odd number of fields.
	l.Info("hello world", "key1", "val1", "key2")
	require.Contains(t, buf.String(), `level=info message="hello world" defaultkey=defaultval key1=val1`)
	buf.Reset()
}

// These test are typically meant to be run with the data race detector.
func TestLoggerConcurrency(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Opts{Writer: buf})

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
