package ups

import (
	"context"
	"net/http"

	tracking "github.com/befship/ups-sdk/internal/generated/tracking"
)

// Re-exported Tracking types.
type (
	TrackingClient          = tracking.ClientWithResponses
	TrackParams             = tracking.GetSingleTrackResponseUsingGETParams
	TrackHTTPResponse       = tracking.GetSingleTrackResponseUsingGETResponse
	TrackApiResponse        = tracking.TrackApiResponse
	ReferenceTrackingParams = tracking.ReferenceTrackingAPIParams
	ShipmentTrackingParams  = tracking.ShipmentTrackingAPIParams
)

// TrackOptions controls Track.
type TrackOptions struct {
	TransID        string
	TransactionSrc string
	Params         *TrackParams
}

// TrackingClient returns a generated Tracking client for the configured environment.
func (c *Client) TrackingClient() (*TrackingClient, error) {
	return tracking.NewClientWithResponses(
		ServerURL(c.cfg.Environment, "tracking"),
		tracking.WithHTTPClient(c.http),
	)
}

func (c *Client) trackingEditors(ctx context.Context, transID, src string) ([]tracking.RequestEditorFn, error) {
	editors, err := c.withAuth(ctx, transID, src)
	if err != nil {
		return nil, err
	}
	out := make([]tracking.RequestEditorFn, len(editors))
	for i, e := range editors {
		fn := e
		out[i] = tracking.RequestEditorFn(fn)
	}
	return out, nil
}

// Track retrieves tracking details for an inquiry number (1Z…, etc.).
func (c *Client) Track(ctx context.Context, inquiryNumber string, opt *TrackOptions) (*TrackApiResponse, error) {
	if opt == nil {
		opt = &TrackOptions{}
	}
	editors, err := c.trackingEditors(ctx, opt.TransID, opt.TransactionSrc)
	if err != nil {
		return nil, err
	}
	cli, err := c.TrackingClient()
	if err != nil {
		return nil, err
	}
	resp, err := cli.GetSingleTrackResponseUsingGETWithResponse(ctx, inquiryNumber, opt.Params, editors...)
	if err != nil {
		return nil, &APIError{Op: "track", Message: "request failed", Err: err}
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		c.InvalidateToken()
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil {
		return nil, responseError("track", resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

// TrackByReference tracks by reference number.
func (c *Client) TrackByReference(ctx context.Context, referenceNumber string, params *ReferenceTrackingParams, opt *TrackOptions) (*TrackApiResponse, error) {
	if opt == nil {
		opt = &TrackOptions{}
	}
	editors, err := c.trackingEditors(ctx, opt.TransID, opt.TransactionSrc)
	if err != nil {
		return nil, err
	}
	cli, err := c.TrackingClient()
	if err != nil {
		return nil, err
	}
	resp, err := cli.ReferenceTrackingAPIWithResponse(ctx, referenceNumber, params, editors...)
	if err != nil {
		return nil, &APIError{Op: "track_reference", Message: "request failed", Err: err}
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		c.InvalidateToken()
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil {
		return nil, responseError("track_reference", resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}
