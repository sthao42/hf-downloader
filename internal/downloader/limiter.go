package downloader

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// NewDownloadHTTPClient creates an HTTP client optimized for high-bandwidth chunked streaming.
// Timeout is intentionally 0 to allow downloading multi-gigabyte files without arbitrary cutoffs.
// Stalls are detected at the socket/inactivity level rather than an artificial global client timeout.
func NewDownloadHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 0,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   32,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// SpeedLimiter provides global bandwidth throttling across all active download chunks
// using a token-bucket rate limiting algorithm.
type SpeedLimiter struct {
	mu         sync.Mutex
	limitBPS   float64   // bytes per second; <= 0 means unlimited
	tokens     float64   // available tokens in bucket
	burst      float64   // max token capacity
	lastRefill time.Time // timestamp of previous token replenish
}

// NewSpeedLimiter creates a new limiter with the given speed limit in MB/s.
// limitMBps <= 0 indicates unlimited speed (no throttling).
func NewSpeedLimiter(limitMBps float64) *SpeedLimiter {
	l := &SpeedLimiter{
		lastRefill: time.Now(),
	}
	l.SetLimit(limitMBps)
	return l
}

// SetLimit dynamically updates the download speed limit in MB/s.
// It is thread-safe and takes effect immediately for ongoing downloads.
func (l *SpeedLimiter) SetLimit(limitMBps float64) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if limitMBps <= 0 {
		l.limitBPS = 0
		l.tokens = 0
		l.burst = 0
		return
	}

	bps := limitMBps * 1024 * 1024
	l.limitBPS = bps

	// Burst bucket capacity: 250ms of throughput, bounded between 256KB and 4MB
	burst := bps * 0.25
	if burst < 256*1024 {
		burst = 256 * 1024
	}
	if burst > 4*1024*1024 {
		burst = 4 * 1024 * 1024
	}
	l.burst = burst
	l.tokens = burst
	l.lastRefill = time.Now()
}

// LimitMBps returns the current speed limit in MB/s (0 = unlimited).
func (l *SpeedLimiter) LimitMBps() float64 {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.limitBPS <= 0 {
		return 0
	}
	return l.limitBPS / (1024 * 1024)
}

// Wait blocks until n bytes of bandwidth quota are available, or ctx is cancelled.
func (l *SpeedLimiter) Wait(ctx context.Context, n int) error {
	if l == nil || n <= 0 {
		return nil
	}

	for {
		l.mu.Lock()
		if l.limitBPS <= 0 {
			l.mu.Unlock()
			return nil
		}

		now := time.Now()
		elapsed := now.Sub(l.lastRefill).Seconds()
		if elapsed > 0 {
			l.tokens += elapsed * l.limitBPS
			if l.tokens > l.burst {
				l.tokens = l.burst
			}
			l.lastRefill = now
		}

		needed := float64(n)
		if l.tokens >= needed {
			l.tokens -= needed
			l.mu.Unlock()
			return nil
		}

		// Calculate sleep time needed to refill required tokens
		deficit := needed - l.tokens
		sleepSecs := deficit / l.limitBPS
		l.mu.Unlock()

		sleepDur := time.Duration(sleepSecs * float64(time.Second))
		if sleepDur < time.Millisecond {
			sleepDur = time.Millisecond
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepDur):
		}
	}
}

// NewReader wraps an io.Reader and throttles read throughput using this limiter.
func (l *SpeedLimiter) NewReader(ctx context.Context, r io.Reader) io.Reader {
	if l == nil {
		return r
	}
	return &limitedReader{
		ctx:     ctx,
		r:       r,
		limiter: l,
	}
}

type limitedReader struct {
	ctx     context.Context
	r       io.Reader
	limiter *SpeedLimiter
}

func (lr *limitedReader) Read(p []byte) (int, error) {
	if lr.limiter == nil {
		return lr.r.Read(p)
	}

	lr.limiter.mu.Lock()
	limitBPS := lr.limiter.limitBPS
	burst := lr.limiter.burst
	lr.limiter.mu.Unlock()

	if limitBPS <= 0 {
		return lr.r.Read(p)
	}

	// Limit buffer slice chunk size so large buffers don't cause huge bursts
	toRead := len(p)
	maxChunk := int(burst)
	if maxChunk > 128*1024 {
		maxChunk = 128 * 1024
	}
	if maxChunk <= 0 {
		maxChunk = 32 * 1024
	}
	if toRead > maxChunk {
		toRead = maxChunk
	}

	n, err := lr.r.Read(p[:toRead])
	if n > 0 {
		if wErr := lr.limiter.Wait(lr.ctx, n); wErr != nil {
			return n, wErr
		}
	}
	return n, err
}
