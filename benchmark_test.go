package logf_test

import (
	"errors"
	"io"
	"testing"

	"github.com/zerodha/logf"
)

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
	logger := logf.New(logf.Opts{Writer: io.Discard, DefaultFields: []interface{}{"component", "logf"}})
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
