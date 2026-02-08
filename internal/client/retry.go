package client

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
	}
}

// RetryWithBackoff retries a request with exponential backoff
func (c *Client) RetryWithBackoff(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	config := DefaultRetryConfig()

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		resp, err := fn()

		// Success - no retry needed
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Save error
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			if resp.StatusCode == http.StatusTooManyRequests {
				// Check for Retry-After header
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, err := strconv.Atoi(retryAfter); err == nil {
						waitDuration := time.Duration(seconds) * time.Second
						if waitDuration > config.MaxBackoff {
							waitDuration = config.MaxBackoff
						}

						fmt.Printf("Rate limit hit. Waiting %s before retry...\n", waitDuration)

						select {
						case <-time.After(waitDuration):
							continue
						case <-ctx.Done():
							return nil, ctx.Err()
						}
					}
				}
			}
		}

		// Last attempt failed
		if attempt == config.MaxRetries {
			break
		}

		// Calculate backoff with exponential increase
		backoff := time.Duration(float64(config.InitialBackoff) * math.Pow(2, float64(attempt)))
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}

		fmt.Printf("Request failed (attempt %d/%d). Retrying in %s...\n", attempt+1, config.MaxRetries+1, backoff)

		// Wait before retry
		select {
		case <-time.After(backoff):
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", config.MaxRetries, lastErr)
}

// ShouldRetry determines if an error is retryable
func ShouldRetry(statusCode int) bool {
	// Retry on rate limits and server errors
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode >= 500 && statusCode < 600 {
		return true
	}
	return false
}
