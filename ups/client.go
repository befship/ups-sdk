// Package ups is a non-official Go SDK for UPS REST APIs.
//
// Generated OpenAPI clients live under internal/generated and are accessed
// through this package. You must supply your own UPS developer OAuth credentials.
package ups

import (
	"context"
	"net/http"
	"time"
)

const (
	// EnvSandbox is the UPS Customer Integration Environment (CIE).
	EnvSandbox = "sandbox"
	// EnvProduction is the UPS production environment.
	EnvProduction = "production"

	// DefaultAPIVersion is the path version used by Rate/Ship wrappers when omitted.
	DefaultAPIVersion = "v2409"
)

// Config holds UPS client credentials and runtime options.
// Pass ClientID / ClientSecret explicitly (from your secret store, DB, flags, etc.).
// The SDK does not read credentials from the environment.
type Config struct {
	ClientID      string // required for authenticated calls
	ClientSecret  string // required for authenticated calls
	AccountNumber string // optional; sent as x-merchant-id on token requests
	Environment   string // "sandbox" (default) or "production"

	HTTPClient     *http.Client
	Timeout        time.Duration // applied when HTTPClient is nil; default 60s
	TransactionSrc string        // optional default transactionSrc header
	MaxRetries     int           // HTTP retries on 429/5xx; default 2
	Logger         Logger        // optional; nil = no logging
}

// Logger is a minimal structured logger hook.
type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

// Client is the hand-written UPS facade (OAuth, retries, API accessors).
type Client struct {
	cfg   Config
	http  *http.Client
	token *tokenCache
}

// New builds a Client. ClientID and ClientSecret are required for authenticated calls.
func New(cfg Config) *Client {
	if cfg.Environment == "" {
		cfg.Environment = EnvSandbox
	}
	if cfg.Environment == "cie" || cfg.Environment == "test" {
		cfg.Environment = EnvSandbox
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	} else if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 2
	}
	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 60 * time.Second
		}
		hc = &http.Client{Timeout: timeout}
	}
	c := &Client{
		cfg:   cfg,
		http:  hc,
		token: newTokenCache(),
	}
	c.ensureRetry()
	return c
}

// Environment returns the configured environment name.
func (c *Client) Environment() string { return c.cfg.Environment }

// HTTPClient returns the underlying HTTP client.
func (c *Client) HTTPClient() *http.Client { return c.http }

// BaseURL returns the UPS host (no /api path) for env.
func BaseURL(env string) string {
	if env == EnvProduction || env == "production" {
		return "https://onlinetools.ups.com"
	}
	return "https://wwwcie.ups.com"
}

// IsProduction reports whether env is production.
func IsProduction(env string) bool {
	return env == EnvProduction || env == "production"
}

func (c *Client) logDebugf(format string, args ...any) {
	if c.cfg.Logger != nil {
		c.cfg.Logger.Debugf(format, args...)
	}
}

func (c *Client) logErrorf(format string, args ...any) {
	if c.cfg.Logger != nil {
		c.cfg.Logger.Errorf(format, args...)
	}
}

// withAuth returns request editors that attach Bearer token and optional headers.
func (c *Client) withAuth(ctx context.Context, transID, transactionSrc string) ([]func(context.Context, *http.Request) error, error) {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	src := transactionSrc
	if src == "" {
		src = c.cfg.TransactionSrc
	}
	if src == "" {
		src = "ups-sdk"
	}
	return []func(context.Context, *http.Request) error{
		BearerToken(tok),
		func(_ context.Context, req *http.Request) error {
			if transID != "" {
				req.Header.Set("transId", transID)
			}
			req.Header.Set("transactionSrc", src)
			return nil
		},
	}, nil
}
