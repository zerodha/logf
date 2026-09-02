package logf_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/zerodha/logf"
)

type benchmarkPayload struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
}

type benchmarkLogValuer struct {
	value string
}

func (v benchmarkLogValuer) LogValue() slog.Value {
	return slog.StringValue(v.value)
}

// ============================================================================
// Logf Benchmarks
// ============================================================================

func BenchmarkNoField(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("hello world")
		}
	})
}

func BenchmarkOneField(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("hello world", "stack", "testing")
		}
	})
}

func BenchmarkOneFieldWithDefaultFields(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, DefaultFields: []any{"component", "logf"}})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("hello world", "stack", "testing")
		}
	})
}

func BenchmarkThreeFields(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("request completed",
				"component", "api", "method", "GET", "bytes", 1<<18,
			)
		}
	})
}

func BenchmarkErrorField(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	b.ReportAllocs()
	b.ResetTimer()

	fakeErr := errors.New("fake error")

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Error("request fields", "error", fakeErr)
		}
	})
}

func BenchmarkHugePayload(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("fetched details",
				"id", 11,
				"title", "perfume Oil",
				"description", "Mega Discount, Impression of A...",
				"price", 13,
				"discountPercentage", 8.4,
				"rating", 4.26,
				"stock", 65,
				"brand", "Impression of Acqua Di Gio",
				"category", "fragrances",
				"thumbnail", "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			)
		}
	})
}

func BenchmarkThreeFields_WithCaller(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, CallerSkipFrameCount: 3, EnableCaller: true})
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("request completed",
				"component", "api", "method", "GET", "bytes", 1<<18,
			)
		}
	})
}

func BenchmarkNoField_WithColor(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, EnableColor: true})
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("hello world")
		}
	})
}

func BenchmarkOneField_WithColor(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, EnableColor: true})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("hello world", "stack", "testing")
		}
	})
}

func BenchmarkThreeFields_WithColor(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, EnableColor: true})
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("request completed",
				"component", "api", "method", "GET", "bytes", 1<<18,
			)
		}
	})
}

func BenchmarkErrorField_WithColor(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, EnableColor: true})
	b.ReportAllocs()
	b.ResetTimer()

	fakeErr := errors.New("fake error")

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Error("request fields", "error", fakeErr)
		}
	})
}

func BenchmarkHugePayload_WithColor(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, EnableColor: true})
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("fetched details",
				"id", 11,
				"title", "perfume Oil",
				"description", "Mega Discount, Impression of A...",
				"price", 13,
				"discountPercentage", 8.4,
				"rating", 4.26,
				"stock", 65,
				"brand", "Impression of Acqua Di Gio",
				"category", "fragrances",
				"thumbnail", "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			)
		}
	})
}

// ============================================================================
// Slog Handler Benchmarks
// ============================================================================

func BenchmarkSlog_NoField(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("hello world")
		}
	})
}

func BenchmarkSlog_OneField(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("hello world", "stack", "testing")
		}
	})
}

func BenchmarkSlog_OneFieldTyped(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("hello world", slog.String("stack", "testing"))
		}
	})
}

func BenchmarkSlog_ThreeFields(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed",
				"component", "api", "method", "GET", "bytes", 1<<18,
			)
		}
	})
}

func BenchmarkSlog_ThreeFieldsTyped(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed",
				slog.String("component", "api"),
				slog.String("method", "GET"),
				slog.Int("bytes", 1<<18),
			)
		}
	})
}

func BenchmarkSlog_ErrorField(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	fakeErr := errors.New("fake error")

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Error("request fields", "error", fakeErr)
		}
	})
}

func BenchmarkSlog_HugePayload(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("fetched details",
				"id", 11,
				"title", "perfume Oil",
				"description", "Mega Discount, Impression of A...",
				"price", 13,
				"discountPercentage", 8.4,
				"rating", 4.26,
				"stock", 65,
				"brand", "Impression of Acqua Di Gio",
				"category", "fragrances",
				"thumbnail", "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			)
		}
	})
}

func BenchmarkSlog_HugePayloadTyped(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("fetched details",
				slog.Int("id", 11),
				slog.String("title", "perfume Oil"),
				slog.String("description", "Mega Discount, Impression of A..."),
				slog.Int("price", 13),
				slog.Float64("discountPercentage", 8.4),
				slog.Float64("rating", 4.26),
				slog.Int("stock", 65),
				slog.String("brand", "Impression of Acqua Di Gio"),
				slog.String("category", "fragrances"),
				slog.String("thumbnail", "https://dummyjson.com/image/i/products/11/thumbnail.jpg"),
			)
		}
	})
}

func BenchmarkSlog_WithAttrs(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	handlerWithAttrs := handler.WithAttrs([]slog.Attr{
		slog.String("component", "api"),
		slog.String("environment", "production"),
	})
	slogger := slog.New(handlerWithAttrs)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed", "method", "GET")
		}
	})
}

func BenchmarkSlog_WithGroup(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	handlerWithGroup := handler.WithGroup("request")
	slogger := slog.New(handlerWithGroup)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed", "method", "GET", "status", 200)
		}
	})
}

func BenchmarkSlog_GroupAttr(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed",
				slog.Group("request",
					slog.String("method", "GET"),
					slog.String("path", "/api/users"),
				),
			)
		}
	})
}

func BenchmarkSlog_NestedGroup(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed",
				slog.Group("http",
					slog.Group("response", slog.Int("status", 200)),
				),
			)
		}
	})
}

func BenchmarkSlog_Disabled(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard, Level: logf.ErrorLevel})
	handler := logf.NewSlogHandler(logger)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("this should be filtered out", "key", "value")
		}
	})
}

// ============================================================================
// Vanilla Slog Benchmarks (for comparison)
// ============================================================================

func BenchmarkVanillaSlog_JSON_NoField(b *testing.B) {
	handler := slog.NewJSONHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("hello world")
		}
	})
}

func BenchmarkVanillaSlog_JSON_OneField(b *testing.B) {
	handler := slog.NewJSONHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("hello world", "stack", "testing")
		}
	})
}

func BenchmarkVanillaSlog_JSON_ThreeFields(b *testing.B) {
	handler := slog.NewJSONHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed",
				"component", "api", "method", "GET", "bytes", 1<<18,
			)
		}
	})
}

func BenchmarkVanillaSlog_JSON_ThreeFieldsTyped(b *testing.B) {
	handler := slog.NewJSONHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed",
				slog.String("component", "api"),
				slog.String("method", "GET"),
				slog.Int("bytes", 1<<18),
			)
		}
	})
}

func BenchmarkVanillaSlog_JSON_HugePayload(b *testing.B) {
	handler := slog.NewJSONHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("fetched details",
				"id", 11,
				"title", "perfume Oil",
				"description", "Mega Discount, Impression of A...",
				"price", 13,
				"discountPercentage", 8.4,
				"rating", 4.26,
				"stock", 65,
				"brand", "Impression of Acqua Di Gio",
				"category", "fragrances",
				"thumbnail", "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			)
		}
	})
}

func BenchmarkVanillaSlog_Text_NoField(b *testing.B) {
	handler := slog.NewTextHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("hello world")
		}
	})
}

func BenchmarkVanillaSlog_Text_ThreeFields(b *testing.B) {
	handler := slog.NewTextHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed",
				"component", "api", "method", "GET", "bytes", 1<<18,
			)
		}
	})
}

func BenchmarkVanillaSlog_Text_HugePayload(b *testing.B) {
	handler := slog.NewTextHandler(io.Discard, nil)
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("fetched details",
				"id", 11,
				"title", "perfume Oil",
				"description", "Mega Discount, Impression of A...",
				"price", 13,
				"discountPercentage", 8.4,
				"rating", 4.26,
				"stock", 65,
				"brand", "Impression of Acqua Di Gio",
				"category", "fragrances",
				"thumbnail", "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			)
		}
	})
}

func BenchmarkVanillaSlog_Text_WithAttrs(b *testing.B) {
	handler := slog.NewTextHandler(io.Discard, nil).WithAttrs([]slog.Attr{
		slog.String("component", "api"),
		slog.String("environment", "production"),
	})
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed", "method", "GET")
		}
	})
}

func BenchmarkSlog_WithAttrsNestedGroup(b *testing.B) {
	logger := logf.New(logf.Opts{Writer: io.Discard})
	handler := logf.NewSlogHandler(logger).WithAttrs([]slog.Attr{
		slog.Group("http",
			slog.Group("response", slog.Int("status", 200)),
		),
	})
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed")
		}
	})
}

func BenchmarkVanillaSlog_Text_WithAttrsNestedGroup(b *testing.B) {
	handler := slog.NewTextHandler(io.Discard, nil).WithAttrs([]slog.Attr{
		slog.Group("http",
			slog.Group("response", slog.Int("status", 200)),
		),
	})
	slogger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			slogger.Info("request completed")
		}
	})
}

func BenchmarkSlogHandle_HotPaths(b *testing.B) {
	ctx := context.Background()
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	b.Run("NoFields", func(b *testing.B) {
		handler := logf.NewSlogHandler(logf.New(logf.Opts{Writer: io.Discard}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("EscapedValue", func(b *testing.B) {
		handler := logf.NewSlogHandler(logf.New(logf.Opts{Writer: io.Discard}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		record.AddAttrs(slog.String("request path", "/users active"))
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("NestedGroup", func(b *testing.B) {
		handler := logf.NewSlogHandler(logf.New(logf.Opts{Writer: io.Discard}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		record.AddAttrs(slog.Group("http",
			slog.Group("response", slog.Int("status", 200)),
		))
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("DefaultFields", func(b *testing.B) {
		handler := logf.NewSlogHandler(logf.New(logf.Opts{
			Writer:        io.Discard,
			DefaultFields: []any{"service", "api", "environment", "production"},
		}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("RFC3339Nano", func(b *testing.B) {
		handler := logf.NewSlogHandler(logf.New(logf.Opts{
			Writer:          io.Discard,
			TimestampFormat: time.RFC3339Nano,
		}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("Caller", func(b *testing.B) {
		handler := logf.NewSlogHandler(logf.New(logf.Opts{
			Writer:       io.Discard,
			EnableCaller: true,
		}))
		var pcs [1]uintptr
		runtime.Callers(1, pcs[:])
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", pcs[0])
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})
}

func BenchmarkVanillaSlog_TextHandle_NestedGroup(b *testing.B) {
	handler := slog.NewTextHandler(io.Discard, nil)
	record := slog.NewRecord(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), slog.LevelInfo, "request completed", 0)
	record.AddAttrs(slog.Group("http",
		slog.Group("response", slog.Int("status", 200)),
	))
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = handler.Handle(ctx, record)
	}
}

func BenchmarkSlogJSON_NoField(b *testing.B) {
	logger := slog.New(logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard})))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("hello world")
		}
	})
}

func BenchmarkSlogJSON_ThreeFields(b *testing.B) {
	logger := slog.New(logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard})))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("request completed", "component", "api", "method", "GET", "bytes", 1<<18)
		}
	})
}

func BenchmarkSlogJSON_HugePayload(b *testing.B) {
	logger := slog.New(logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard})))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("fetched details",
				"id", 11,
				"title", "perfume Oil",
				"description", "Mega Discount, Impression of A...",
				"price", 13,
				"discountPercentage", 8.4,
				"rating", 4.26,
				"stock", 65,
				"brand", "Impression of Acqua Di Gio",
				"category", "fragrances",
				"thumbnail", "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			)
		}
	})
}

func BenchmarkSlogJSON_WithAttrs(b *testing.B) {
	handler := logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard})).WithAttrs([]slog.Attr{
		slog.String("component", "api"),
		slog.String("environment", "production"),
	})
	logger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("request completed", "method", "GET")
		}
	})
}

func BenchmarkSlogJSON_NestedGroup(b *testing.B) {
	logger := slog.New(logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard})))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("request completed",
				slog.Group("http",
					slog.Group("response", slog.Int("status", 200)),
				),
			)
		}
	})
}

func BenchmarkSlogJSON_StructuredAny(b *testing.B) {
	logger := slog.New(logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard})))
	payload := benchmarkPayload{ID: 7, Name: "worker", Labels: []string{"api", "production"}}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("worker ready", "payload", payload)
		}
	})
}

func BenchmarkSlogJSON_DisabledLogValuer(b *testing.B) {
	logger := slog.New(logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard, Level: logf.ErrorLevel})))
	ctx := context.Background()
	attr := slog.Any("value", benchmarkLogValuer{value: "deferred"})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "disabled", attr)
		}
	})
}

func BenchmarkVanillaSlog_JSON_WithAttrs(b *testing.B) {
	handler := slog.NewJSONHandler(io.Discard, nil).WithAttrs([]slog.Attr{
		slog.String("component", "api"),
		slog.String("environment", "production"),
	})
	logger := slog.New(handler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("request completed", "method", "GET")
		}
	})
}

func BenchmarkVanillaSlog_JSON_NestedGroup(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("request completed",
				slog.Group("http",
					slog.Group("response", slog.Int("status", 200)),
				),
			)
		}
	})
}

func BenchmarkVanillaSlog_JSON_StructuredAny(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	payload := benchmarkPayload{ID: 7, Name: "worker", Labels: []string{"api", "production"}}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.Info("worker ready", "payload", payload)
		}
	})
}

func BenchmarkVanillaSlog_JSON_DisabledLogValuer(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := context.Background()
	attr := slog.Any("value", benchmarkLogValuer{value: "deferred"})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "disabled", attr)
		}
	})
}

func BenchmarkSlogJSONHandle_HotPaths(b *testing.B) {
	ctx := context.Background()
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	b.Run("NoFields", func(b *testing.B) {
		handler := logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("NestedGroup", func(b *testing.B) {
		handler := logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		record.AddAttrs(slog.Group("http", slog.Group("response", slog.Int("status", 200))))
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("DefaultFields", func(b *testing.B) {
		handler := logf.NewSlogJSONHandler(logf.New(logf.Opts{
			Writer:        io.Discard,
			DefaultFields: []any{"service", "api", "environment", "production"},
		}))
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("Caller", func(b *testing.B) {
		handler := logf.NewSlogJSONHandler(logf.New(logf.Opts{Writer: io.Discard, EnableCaller: true}))
		var pcs [1]uintptr
		runtime.Callers(1, pcs[:])
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", pcs[0])
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})
}

func BenchmarkVanillaSlogJSONHandle_HotPaths(b *testing.B) {
	ctx := context.Background()
	recordTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	b.Run("NoFields", func(b *testing.B) {
		handler := slog.NewJSONHandler(io.Discard, nil)
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("NestedGroup", func(b *testing.B) {
		handler := slog.NewJSONHandler(io.Discard, nil)
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		record.AddAttrs(slog.Group("http", slog.Group("response", slog.Int("status", 200))))
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("DefaultFields", func(b *testing.B) {
		handler := slog.NewJSONHandler(io.Discard, nil).WithAttrs([]slog.Attr{
			slog.String("service", "api"),
			slog.String("environment", "production"),
		})
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", 0)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})

	b.Run("Caller", func(b *testing.B) {
		handler := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{AddSource: true})
		var pcs [1]uintptr
		runtime.Callers(1, pcs[:])
		record := slog.NewRecord(recordTime, slog.LevelInfo, "request completed", pcs[0])
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = handler.Handle(ctx, record)
		}
	})
}
