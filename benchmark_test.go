package logf_test

import (
	"errors"
	"io"
	"testing"

	"github.com/zerodha/logf"
)

func BenchmarkNoField(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("hello world")
		}
	})
}

func BenchmarkOneField(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithFields(logf.Fields{"stack": "testing"}).Info("hello world")
		}
	})
}

func BenchmarkThreeFields(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithFields(logf.Fields{
				"component": "api",
				"method":    "GET",
				"bytes":     1 << 18,
			}).Info("request completed")
		}
	})
}

func BenchmarkErrorField(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()

	fakeErr := errors.New("fake error")

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithError(fakeErr).Error("request failed")
		}
	})
}

func BenchmarkHugePayload(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithFields(logf.Fields{
				"id":                 11,
				"title":              "perfume Oil",
				"description":        "Mega Discount, Impression of A...",
				"price":              13,
				"discountPercentage": 8.4,
				"rating":             4.26,
				"stock":              65,
				"brand":              "Impression of Acqua Di Gio",
				"category":           "fragrances",
				"thumbnail":          "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			}).Info("fetched details")
		}
	})
}

func BenchmarkThreeFields_WithCaller(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetCallerFrame(true, 3)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithFields(logf.Fields{
				"component": "api",
				"method":    "GET",
				"bytes":     1 << 18,
			}).Info("request completed")
		}
	})
}

func BenchmarkNoField_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.Info("hello world")
		}
	})
}

func BenchmarkOneField_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithFields(logf.Fields{"stack": "testing"}).Info("hello world")
		}
	})
}

func BenchmarkThreeFields_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithFields(logf.Fields{
				"component": "api",
				"method":    "GET",
				"bytes":     1 << 18,
			}).Info("request completed")
		}
	})
}

func BenchmarkErrorField_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()

	fakeErr := errors.New("fake error")

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithError(fakeErr).Error("request failed")
		}
	})
}

func BenchmarkHugePayload_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			logger.WithFields(logf.Fields{
				"id":                 11,
				"title":              "perfume Oil",
				"description":        "Mega Discount, Impression of A...",
				"price":              13,
				"discountPercentage": 8.4,
				"rating":             4.26,
				"stock":              65,
				"brand":              "Impression of Acqua Di Gio",
				"category":           "fragrances",
				"thumbnail":          "https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			}).Info("fetched details")
		}
	})
}
