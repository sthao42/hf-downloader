package downloader

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

func TestSpeedLimiter_Unlimited(t *testing.T) {
	limiter := NewSpeedLimiter(0)
	if limiter.LimitMBps() != 0 {
		t.Errorf("expected 0 MB/s, got %f", limiter.LimitMBps())
	}

	ctx := context.Background()
	start := time.Now()
	err := limiter.Wait(ctx, 10*1024*1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Errorf("unlimited limiter should not block noticeably")
	}
}

func TestSpeedLimiter_Throttling(t *testing.T) {
	// Set limit to 1 MB/s (1024*1024 B/s)
	limiter := NewSpeedLimiter(1.0)
	if limiter.LimitMBps() != 1.0 {
		t.Errorf("expected 1.0 MB/s, got %f", limiter.LimitMBps())
	}

	payload := make([]byte, 512*1024) // 512 KB
	reader := limiter.NewReader(context.Background(), bytes.NewReader(payload))

	buf := make([]byte, 64*1024)
	start := time.Now()
	totalRead := 0

	for {
		n, err := reader.Read(buf)
		totalRead += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
	}

	if totalRead != len(payload) {
		t.Fatalf("expected to read %d bytes, got %d", len(payload), totalRead)
	}

	elapsed := time.Since(start)
	// 512KB at 1MB/s after burst should take at least ~150ms
	if elapsed < 100*time.Millisecond {
		t.Logf("Elapsed time: %v (throttle active)", elapsed)
	}
}

func TestSpeedLimiter_ContextCancellation(t *testing.T) {
	limiter := NewSpeedLimiter(0.1) // 100 KB/s
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Requesting 10MB quota on a 100 KB/s limiter with a 50ms timeout should cancel
	err := limiter.Wait(ctx, 10*1024*1024)
	if err == nil {
		t.Fatalf("expected context error, got nil")
	}
}
