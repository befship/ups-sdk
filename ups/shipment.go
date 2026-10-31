package ups

import (
	"context"
	"net/http"

	shipping "github.com/befship/ups-sdk/internal/generated/shipping"
)

// Re-exported Shipping types.
type (
	ShippingClient               = shipping.ClientWithResponses
	ShipmentParams               = shipping.ShipmentOpParams
	ShipmentRequest              = shipping.ShipmentOpJSONRequestBody
	ShipmentHTTPResponse         = shipping.ShipmentOpResponse
	SHIPRequestWrapper           = shipping.SHIPRequestWrapper
	SHIPResponseWrapper          = shipping.SHIPResponseWrapper
	VoidShipmentParams           = shipping.VoidShipmentOpParams
	VoidShipmentHTTPResponse     = shipping.VoidShipmentOpResponse
	VOIDSHIPMENTResponseWrapper  = shipping.VOIDSHIPMENTResponseWrapper
	LabelRecoveryParams          = shipping.LabelRecoveryOpParams
	LabelRecoveryRequest         = shipping.LabelRecoveryOpJSONRequestBody
	LABELRECOVERYRequestWrapper  = shipping.LABELRECOVERYRequestWrapper
	LABELRECOVERYResponseWrapper = shipping.LABELRECOVERYResponseWrapper
)

// ShipmentOptions controls CreateShipment.
type ShipmentOptions struct {
	Version                     string // default DefaultAPIVersion
	AdditionalAddressValidation string
	TransID                     string
	TransactionSrc              string
}

// VoidOptions controls VoidShipment.
type VoidOptions struct {
	Version        string
	TransID        string
	TransactionSrc string
	Params         *VoidShipmentParams
}

// LabelRecoveryOptions controls LabelRecovery.
type LabelRecoveryOptions struct {
	Version        string
	TransID        string
	TransactionSrc string
	Params         *LabelRecoveryParams
}

// ShippingClient returns a generated Shipping client for the configured environment.
func (c *Client) ShippingClient() (*ShippingClient, error) {
	return shipping.NewClientWithResponses(
		ServerURL(c.cfg.Environment, "shipping"),
		shipping.WithHTTPClient(c.http),
	)
}

func (c *Client) shippingEditors(ctx context.Context, transID, src string) ([]shipping.RequestEditorFn, error) {
	editors, err := c.withAuth(ctx, transID, src)
	if err != nil {
		return nil, err
	}
	out := make([]shipping.RequestEditorFn, len(editors))
	for i, e := range editors {
		fn := e
		out[i] = shipping.RequestEditorFn(fn)
	}
	return out, nil
}

// CreateShipment creates a shipment and returns label payload (base64 in response GraphicImage when requested).
func (c *Client) CreateShipment(ctx context.Context, body ShipmentRequest, opt *ShipmentOptions) (*SHIPResponseWrapper, error) {
	if opt == nil {
		opt = &ShipmentOptions{}
	}
	version := opt.Version
	if version == "" {
		version = DefaultAPIVersion
	}
	editors, err := c.shippingEditors(ctx, opt.TransID, opt.TransactionSrc)
	if err != nil {
		return nil, err
	}
	cli, err := c.ShippingClient()
	if err != nil {
		return nil, err
	}
	var params *ShipmentParams
	if opt.AdditionalAddressValidation != "" {
		v := opt.AdditionalAddressValidation
		params = &ShipmentParams{Additionaladdressvalidation: &v}
	}
	resp, err := cli.ShipmentOpWithResponse(ctx, version, params, body, editors...)
	if err != nil {
		return nil, &APIError{Op: "ship", Message: "request failed", Err: err}
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		c.InvalidateToken()
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil {
		return nil, responseError("ship", resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

// VoidShipment cancels/voids a shipment by shipment identification number.
func (c *Client) VoidShipment(ctx context.Context, shipmentID string, opt *VoidOptions) (*VOIDSHIPMENTResponseWrapper, error) {
	if opt == nil {
		opt = &VoidOptions{}
	}
	version := opt.Version
	if version == "" {
		version = DefaultAPIVersion
	}
	editors, err := c.shippingEditors(ctx, opt.TransID, opt.TransactionSrc)
	if err != nil {
		return nil, err
	}
	cli, err := c.ShippingClient()
	if err != nil {
		return nil, err
	}
	resp, err := cli.VoidShipmentOpWithResponse(ctx, version, shipmentID, opt.Params, editors...)
	if err != nil {
		return nil, &APIError{Op: "void", Message: "request failed", Err: err}
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		c.InvalidateToken()
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil {
		return nil, responseError("void", resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

// LabelRecovery recovers a previously generated label.
func (c *Client) LabelRecovery(ctx context.Context, body LabelRecoveryRequest, opt *LabelRecoveryOptions) (*LABELRECOVERYResponseWrapper, error) {
	if opt == nil {
		opt = &LabelRecoveryOptions{}
	}
	version := opt.Version
	if version == "" {
		version = DefaultAPIVersion
	}
	editors, err := c.shippingEditors(ctx, opt.TransID, opt.TransactionSrc)
	if err != nil {
		return nil, err
	}
	cli, err := c.ShippingClient()
	if err != nil {
		return nil, err
	}
	resp, err := cli.LabelRecoveryOpWithResponse(ctx, version, opt.Params, body, editors...)
	if err != nil {
		return nil, &APIError{Op: "label_recovery", Message: "request failed", Err: err}
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		c.InvalidateToken()
	}
	if resp.StatusCode() >= 300 || resp.JSON200 == nil {
		return nil, responseError("label_recovery", resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}
