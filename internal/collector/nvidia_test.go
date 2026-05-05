package collector

import (
	"context"
	"testing"
	"time"
)

func BenchmarkFetchNVML_Cold(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulates cold start each tick (old implementation)
		n := NewNVMLCollector()
		n.ensureInit(context.Background())
		_, _ = n.Fetch(context.Background())
		n.Close()
	}
}

func BenchmarkFetchNVML_Warm(b *testing.B) {
	n := NewNVMLCollector()
	n.ensureInit(context.Background())
	defer n.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = n.Fetch(context.Background())
	}
}

func BenchmarkFetchNVML_PerTick(b *testing.B) {
	n := NewNVMLCollector()
	n.ensureInit(context.Background())
	defer n.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, _ = n.Fetch(ctx)
		cancel()
	}
}
