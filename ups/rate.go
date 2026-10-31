package ups

import (
	"context"
	"net/http"

	rating "github.com/Bef-Ship/ups-sdk/internal/generated/rating"
)

// Re-exported Rating types (import path stays github.com/Bef-Ship/ups-sdk/ups).
type (
	RateClient          = rating.ClientWithResponses
	RateParams          = rating.RateOpParams
	RateRequest         = rating.RateOpJSONRequestBody
	RateHTTPResponse    = rating.RateOpResponse
	RATERequestWrapper  = rating.RATERequestWrapper
	RATEResponseWrapper = rating.RATEResponseWrapper
)

// RateOptions controls the Rate API call.
type RateOptions struct {
	Version        string // default DefaultAPIVersion
	RequestOption  string // Rate | Shop | Ratetimeintransit | Shoptimeintransit; default "Rate"
	AdditionalInfo string // e.g. "timeintransit"
	TransID        string
	TransactionSrc string
}

// RatingClient returns a generated Rating client pointed at the configured environment.
func (c *Client) RatingClient() (*RateClient, error) {
	return rating.NewClientWithResponses(
		ServerURL(c.cfg.Environment, "rating"),
		rating.WithHTTPClient(c.http),
	)
}

// Rate quotes a shipment (supports multi-package via RateRequest.RateRequest.Shipment.Package).
func (c *Client) Rate(ctx context.Context, body RateRequest, opt *RateOptions) (*RATEResponseWrapper, error) {
	if opt == nil {
		opt = &RateOptions{}
	}
	version := opt.Version
	if version == "" {
		version = DefaultAPIVersion
	}
	reqOpt := opt.RequestOption
	if reqOpt == "" {
		reqOpt = "Rate"
	}

	editors, err := c.withAuth(ctx, opt.TransID, opt.TransactionSrc)
	if err != nil {
		return nil, err
	}

	cli, err := c.RatingClient()
	if err != nil {
		return nil, err
	}

	var params *RateParams
	if opt.AdditionalInfo != "" {
		ai := opt.AdditionalInfo
		params = &RateParams{Additionalinfo: &ai}
	}

	// Convert editors to rating.RequestEditorFn
	reqEditors := make([]rating.RequestEditorFn, len(editors))
	for i, e := range editors {
		fn := e
		reqEditors[i] = rating.RequestEditorFn(fn)
	}

	resp, err := cli.RateOpWithResponse(ctx, version, reqOpt, params, body, reqEditors...)
	if err != nil {
		return nil, &APIError{Op: "rate", Message: "request failed", Err: err}
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		c.InvalidateToken()
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil {
		return nil, responseError("rate", resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}
