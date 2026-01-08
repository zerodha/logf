package logf_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/zerodha/logf"
)

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
