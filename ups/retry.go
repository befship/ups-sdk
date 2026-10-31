package ups

import (
	"math"
	"net/http"
	"time"
)

// roundTripper with exponential backoff for 429 / 5xx.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	log        Logger
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	var resp *http.Response
	var err error
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		if attempt > 0 {
			// rewind body if possible
			if req.GetBody != nil {
				body, gerr := req.GetBody()
				if gerr != nil {
					return resp, gerr
				}
				req.Body = body
			}
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * 200 * time.Millisecond
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
			if t.log != nil {
				t.log.Debugf("ups: retry attempt=%d delay=%s", attempt, delay)
			}
			timer := time.NewTimer(delay)
			select {
			case <-req.Context().Done():
				timer.Stop()
				return nil, req.Context().Err()
			case <-timer.C:
			}
		}
		resp, err = base.RoundTrip(req)
		if err != nil {
			if attempt == t.maxRetries {
				return nil, err
			}
			continue
		}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return resp, nil
		}
		if attempt == t.maxRetries {
			return resp, nil
		}
		_ = resp.Body.Close()
	}
	return resp, err
}

// wrapHTTP applies retry transport when MaxRetries > 0 and the client uses the default transport path.
func (c *Client) ensureRetry() {
	if c.cfg.MaxRetries <= 0 {
		return
	}
	if _, ok := c.http.Transport.(*retryTransport); ok {
		return
	}
	// Only wrap if caller did not set a custom transport on a shared client they own carefully.
	if c.cfg.HTTPClient != nil {
		return
	}
	c.http.Transport = &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: c.cfg.MaxRetries,
		log:        c.cfg.Logger,
	}
}
