package provider

import (
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// retryTransport wraps an http.RoundTripper and implements exponential backoff
// retry logic for transient errors, particularly 429 Too Many Requests.
type retryTransport struct {
	rt              http.RoundTripper
	maxRetryTimeout time.Duration // Maximum total time to spend retrying (default 5 minutes)
	initialBackoff  time.Duration // Initial backoff duration (default 1 second)
	maxBackoff      time.Duration // Maximum backoff duration (default 60 seconds)
	backoffFactor   float64       // Multiplier for exponential backoff (default 2.0)
}

// retryTransportOption is a functional option for configuring retryTransport.
type retryTransportOption func(*retryTransport)

// WithMaxRetryTimeout sets the maximum total time to spend retrying.
func WithMaxRetryTimeout(d time.Duration) retryTransportOption {
	return func(rt *retryTransport) {
		rt.maxRetryTimeout = d
	}
}

// WithInitialBackoff sets the initial backoff duration.
func WithInitialBackoff(d time.Duration) retryTransportOption {
	return func(rt *retryTransport) {
		rt.initialBackoff = d
	}
}

// WithMaxBackoff sets the maximum backoff duration.
func WithMaxBackoff(d time.Duration) retryTransportOption {
	return func(rt *retryTransport) {
		rt.maxBackoff = d
	}
}

// WithBackoffFactor sets the multiplier for exponential backoff.
func WithBackoffFactor(f float64) retryTransportOption {
	return func(rt *retryTransport) {
		rt.backoffFactor = f
	}
}

// newRetryTransport creates a new retryTransport wrapping the given RoundTripper.
// Default configuration:
//   - maxRetryTimeout: 5 minutes
//   - initialBackoff: 1 second
//   - maxBackoff: 60 seconds
//   - backoffFactor: 2.0
func newRetryTransport(rt http.RoundTripper, opts ...retryTransportOption) *retryTransport {
	if rt == nil {
		rt = http.DefaultTransport
	}

	t := &retryTransport{
		rt:              rt,
		maxRetryTimeout: 5 * time.Minute,
		initialBackoff:  1 * time.Second,
		maxBackoff:      60 * time.Second,
		backoffFactor:   2.0,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RoundTrip implements http.RoundTripper with retry logic for 429 responses.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	deadline := time.Now().Add(t.maxRetryTimeout)
	attempt := 0
	backoff := t.initialBackoff

	for {
		// Clone the request for retry (body needs to be re-readable)
		reqCopy := req.Clone(req.Context())

		// If the request has a body, we need to handle it specially
		if req.Body != nil && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			reqCopy.Body = body
		}

		resp, err := t.rt.RoundTrip(reqCopy)

		// If there's a network error or the request was successful, return immediately
		if err != nil {
			return nil, err
		}

		// Check if we should retry
		if !t.shouldRetry(resp) {
			return resp, nil
		}

		// Check if we've exceeded the deadline
		if time.Now().After(deadline) {
			log.Printf("pingoneprovisioning: retry timeout exceeded after %d attempts for %s %s",
				attempt+1, req.Method, req.URL.String())
			return resp, nil
		}

		// Calculate sleep duration
		sleepDuration := t.calculateBackoff(resp, backoff)

		// Don't sleep longer than the remaining time
		remaining := time.Until(deadline)
		if sleepDuration > remaining {
			sleepDuration = remaining
		}

		// Drain and close the response body before retrying
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		log.Printf("pingoneprovisioning: received %d for %s %s, retrying in %s (attempt %d)",
			resp.StatusCode, req.Method, req.URL.String(), sleepDuration.Round(time.Millisecond), attempt+1)

		// Sleep before retry
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(sleepDuration):
		}

		// Increase backoff for next iteration
		backoff = t.nextBackoff(backoff)
		attempt++
	}
}

// shouldRetry determines if the response indicates a retryable error.
func (t *retryTransport) shouldRetry(resp *http.Response) bool {
	switch resp.StatusCode {
	case http.StatusTooManyRequests: // 429
		return true
	case http.StatusServiceUnavailable: // 503
		return true
	case http.StatusGatewayTimeout: // 504
		return true
	case http.StatusBadGateway: // 502
		return true
	default:
		return false
	}
}

// calculateBackoff determines how long to wait before the next retry.
// It respects the Retry-After header if present, otherwise uses exponential backoff with jitter.
func (t *retryTransport) calculateBackoff(resp *http.Response, currentBackoff time.Duration) time.Duration {
	// Check for Retry-After header
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		// Try parsing as seconds first
		if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
			duration := time.Duration(seconds) * time.Second
			// Cap at maxBackoff
			if duration > t.maxBackoff {
				duration = t.maxBackoff
			}
			return duration
		}

		// Try parsing as HTTP date
		if retryTime, err := http.ParseTime(retryAfter); err == nil {
			duration := time.Until(retryTime)
			if duration < 0 {
				duration = t.initialBackoff
			}
			if duration > t.maxBackoff {
				duration = t.maxBackoff
			}
			return duration
		}
	}

	// Use exponential backoff with jitter
	// Add jitter: random value between 0 and 25% of the backoff
	jitter := time.Duration(rand.Float64() * 0.25 * float64(currentBackoff))
	return currentBackoff + jitter
}

// nextBackoff calculates the next backoff duration using exponential growth.
func (t *retryTransport) nextBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * t.backoffFactor)
	if next > t.maxBackoff {
		next = t.maxBackoff
	}
	return next
}

// retryRoundTripper wraps an existing transport to add retry capability while
// preserving other transport behaviors (like OAuth2).
type retryRoundTripper struct {
	base            http.RoundTripper
	maxRetryTimeout time.Duration
	initialBackoff  time.Duration
	maxBackoff      time.Duration
	backoffFactor   float64
}

// newRetryRoundTripper creates a retry wrapper for any RoundTripper.
func newRetryRoundTripper(base http.RoundTripper) *retryRoundTripper {
	return &retryRoundTripper{
		base:            base,
		maxRetryTimeout: 5 * time.Minute,
		initialBackoff:  1 * time.Second,
		maxBackoff:      60 * time.Second,
		backoffFactor:   2.0,
	}
}

// RoundTrip implements http.RoundTripper with retry logic.
func (r *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	deadline := time.Now().Add(r.maxRetryTimeout)
	attempt := 0
	backoff := r.initialBackoff

	for {
		resp, err := r.base.RoundTrip(req)

		if err != nil {
			return nil, err
		}

		if !r.shouldRetry(resp) {
			return resp, nil
		}

		if time.Now().After(deadline) {
			log.Printf("pingoneprovisioning: retry timeout exceeded after %d attempts for %s %s",
				attempt+1, req.Method, req.URL.String())
			return resp, nil
		}

		sleepDuration := r.calculateBackoff(resp, backoff)

		remaining := time.Until(deadline)
		if sleepDuration > remaining {
			sleepDuration = remaining
		}

		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		log.Printf("pingoneprovisioning: received %d for %s %s, retrying in %s (attempt %d)",
			resp.StatusCode, req.Method, req.URL.String(), sleepDuration.Round(time.Millisecond), attempt+1)

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(sleepDuration):
		}

		backoff = r.nextBackoff(backoff)
		attempt++
	}
}

func (r *retryRoundTripper) shouldRetry(resp *http.Response) bool {
	switch resp.StatusCode {
	case http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusBadGateway:
		return true
	default:
		return false
	}
}

func (r *retryRoundTripper) calculateBackoff(resp *http.Response, currentBackoff time.Duration) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
			duration := time.Duration(seconds) * time.Second
			if duration > r.maxBackoff {
				duration = r.maxBackoff
			}
			return duration
		}
		if retryTime, err := http.ParseTime(retryAfter); err == nil {
			duration := time.Until(retryTime)
			if duration < 0 {
				duration = r.initialBackoff
			}
			if duration > r.maxBackoff {
				duration = r.maxBackoff
			}
			return duration
		}
	}

	// Exponential backoff with jitter
	jitter := time.Duration(rand.Float64() * 0.25 * float64(currentBackoff))
	return currentBackoff + jitter
}

func (r *retryRoundTripper) nextBackoff(current time.Duration) time.Duration {
	next := time.Duration(math.Min(float64(current)*r.backoffFactor, float64(r.maxBackoff)))
	return next
}
