package fetcher

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type Fetcher struct {
	client  *http.Client
	retries int
}

func New(timeout time.Duration, retries int) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        2,
				MaxIdleConnsPerHost: 1,
				IdleConnTimeout:     30 * time.Second,
				DialContext: (&net.Dialer{
					Timeout:   3 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			},
		},
		retries: retries,
	}
}

// Fetch gets URL body with retry + exponential backoff.
func (f *Fetcher) Fetch(ctx context.Context, url string) (string, error) {
	var lastErr error
	backoff := 100 * time.Millisecond
	for attempt := 0; attempt < f.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}
		var body string
		body, lastErr = f.do(ctx, url)
		if lastErr == nil {
			return body, nil
		}
	}
	return "", fmt.Errorf("fetch failed after %d retries: %w", f.retries, lastErr)
}

func (f *Fetcher) do(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	return string(b), nil
}
