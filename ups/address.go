package ups

import (
	"context"
	"net/http"

	av "github.com/befship/ups-sdk/internal/generated/addressvalidation"
)

// Re-exported Address Validation types.
type (
	AddressValidationClient       = av.ClientWithResponses
	AddressValidationParams       = av.AddressValidationParams
	AddressValidationRequest      = av.AddressValidationJSONRequestBody
	XAVRequestWrapper             = av.XAVRequestWrapper
	XAVResponseWrapper            = av.XAVResponseWrapper
	AddressValidationHTTPResponse = av.AddressValidationResponse
)

// AddressValidationOptions controls ValidateAddress.
type AddressValidationOptions struct {
	Version        string // default "v2"
	RequestOption  int    // 1=validation, 2=classification, 3=both; default 1
	TransID        string
	TransactionSrc string
	Params         *AddressValidationParams
}

// AddressValidationClient returns a generated Address Validation client.
func (c *Client) AddressValidationClient() (*AddressValidationClient, error) {
	return av.NewClientWithResponses(
		ServerURL(c.cfg.Environment, "addressvalidation"),
		av.WithHTTPClient(c.http),
	)
}

// ValidateAddress validates / classifies an address.
func (c *Client) ValidateAddress(ctx context.Context, body AddressValidationRequest, opt *AddressValidationOptions) (*XAVResponseWrapper, error) {
	if opt == nil {
		opt = &AddressValidationOptions{}
	}
	version := opt.Version
	if version == "" {
		version = "v2"
	}
	reqOpt := opt.RequestOption
	if reqOpt == 0 {
		reqOpt = 1
	}
	editors, err := c.withAuth(ctx, opt.TransID, opt.TransactionSrc)
	if err != nil {
		return nil, err
	}
	cli, err := c.AddressValidationClient()
	if err != nil {
		return nil, err
	}
	reqEditors := make([]av.RequestEditorFn, len(editors))
	for i, e := range editors {
		fn := e
		reqEditors[i] = av.RequestEditorFn(fn)
	}
	resp, err := cli.AddressValidationWithResponse(ctx, version, reqOpt, opt.Params, body, reqEditors...)
	if err != nil {
		return nil, &APIError{Op: "address_validation", Message: "request failed", Err: err}
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		c.InvalidateToken()
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil {
		return nil, responseError("address_validation", resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}
