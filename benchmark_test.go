package logf_test

import (
	"errors"
	"io"
	"testing"

	"github.com/mr-karan/logf"
)

func BenchmarkNoField(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("hello world")
	}
}

func BenchmarkOneField(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logf.Fields{"stack": "testing"}).Info("hello world")
	}
}

func BenchmarkThreeFields(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(logf.Fields{
			"component": "api",
			"method":    "GET",
			"bytes":     1 << 18,
		}).Info("request completed")
	}
}

func BenchmarkThreeFields_WithCaller(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetCallerFrame(true, 3)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(logf.Fields{
			"component": "api",
			"method":    "GET",
			"bytes":     1 << 18,
		}).Info("request completed")
	}
}

func BenchmarkErrorField(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()

	fakeErr := errors.New("fake error")

	for i := 0; i < b.N; i++ {
		logger.WithError(fakeErr).Error("request failed")
	}
}

func BenchmarkHugePayload(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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
			"images": []string{
				"https://dummyjson.com/image/i/products/11/1.jpg",
				"https://dummyjson.com/image/i/products/11/2.jpg",
				"https://dummyjson.com/image/i/products/11/3.jpg",
				"https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			},
		}).Info("fetched details")
	}
}

func BenchmarkNoField_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("hello world")
	}
}

func BenchmarkOneField_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logf.Fields{"stack": "testing"}).Info("hello world")
	}
}

func BenchmarkThreeFields_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(logf.Fields{
			"component": "api",
			"method":    "GET",
			"bytes":     1 << 18,
		}).Info("request completed")
	}
}

func BenchmarkHugePayload_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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
			"images": []string{
				"https://dummyjson.com/image/i/products/11/1.jpg",
				"https://dummyjson.com/image/i/products/11/2.jpg",
				"https://dummyjson.com/image/i/products/11/3.jpg",
				"https://dummyjson.com/image/i/products/11/thumbnail.jpg",
			},
		}).Info("fetched details")
	}
}

func BenchmarkErrorField_WithColor(b *testing.B) {
	logger := logf.New()
	logger.SetWriter(io.Discard)
	logger.SetColorOutput(true)
	b.ReportAllocs()
	b.ResetTimer()

	fakeErr := errors.New("fake error")

	for i := 0; i < b.N; i++ {
		logger.WithError(fakeErr).Error("request failed")
	}
}
