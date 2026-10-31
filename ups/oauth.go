package ups

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	oauthgen "github.com/Bef-Ship/ups-sdk/internal/generated/oauthclientcredentials"
)

type tokenCache struct {
	mu    sync.Mutex
	token string
	exp   time.Time
}

func newTokenCache() *tokenCache { return &tokenCache{} }

func (t *tokenCache) get() (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.token == "" || time.Now().After(t.exp) {
		return "", false
	}
	return t.token, true
}

func (t *tokenCache) set(token string, expiresInSec int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.token = token
	skew := 60 * time.Second
	if expiresInSec <= 0 {
		expiresInSec = 3300
	}
	d := time.Duration(expiresInSec)*time.Second - skew
	if d < time.Minute {
		d = time.Minute
	}
	t.exp = time.Now().Add(d)
}

func (t *tokenCache) invalidate() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.token = ""
	t.exp = time.Time{}
}

// BasicAuth returns a request editor that sets HTTP Basic auth (OAuth token call).
func BasicAuth(clientID, clientSecret string) oauthgen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)))
		return nil
	}
}

// BearerToken sets Authorization: Bearer on the request.
func BearerToken(token string) func(context.Context, *http.Request) error {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

// AccessToken returns a cached or freshly fetched OAuth access token (client credentials).
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.cfg.ClientID == "" || c.cfg.ClientSecret == "" {
		return "", fmt.Errorf("ups: ClientID and ClientSecret are required")
	}
	if tok, ok := c.token.get(); ok {
		return tok, nil
	}
	return c.fetchToken(ctx)
}

// InvalidateToken clears the cached access token (forces refresh on next call).
func (c *Client) InvalidateToken() {
	c.token.invalidate()
}

func (c *Client) fetchToken(ctx context.Context) (string, error) {
	rc, err := oauthgen.NewClientWithResponses(
		ServerURL(c.cfg.Environment, "oauthclientcredentials"),
		oauthgen.WithHTTPClient(c.http),
	)
	if err != nil {
		return "", fmt.Errorf("ups: oauth client: %w", err)
	}

	var params *oauthgen.CreateTokenParams
	if c.cfg.AccountNumber != "" {
		mid := c.cfg.AccountNumber
		params = &oauthgen.CreateTokenParams{XMerchantId: &mid}
	}

	c.logDebugf("ups: fetching OAuth token (env=%s)", c.cfg.Environment)
	resp, err := rc.CreateTokenWithFormdataBodyWithResponse(ctx, params, oauthgen.CreateTokenFormdataRequestBody{
		GrantType: "client_credentials",
	}, BasicAuth(c.cfg.ClientID, c.cfg.ClientSecret))
	if err != nil {
		return "", &APIError{Op: "oauth", Message: "token request failed", Err: err}
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil || resp.JSON200.AccessToken == nil || *resp.JSON200.AccessToken == "" {
		return "", &APIError{
			Op:         "oauth",
			StatusCode: resp.StatusCode(),
			Body:       string(resp.Body),
			Message:    "token request rejected",
		}
	}
	tok := *resp.JSON200.AccessToken
	sec := 0
	if resp.JSON200.ExpiresIn != nil {
		sec, _ = strconv.Atoi(*resp.JSON200.ExpiresIn)
	}
	c.token.set(tok, sec)
	return tok, nil
}

// AuthedEditors returns Bearer (+ optional headers) request editors after ensuring a token.
func (c *Client) AuthedEditors(ctx context.Context, transID, transactionSrc string) ([]func(context.Context, *http.Request) error, error) {
	return c.withAuth(ctx, transID, transactionSrc)
}
