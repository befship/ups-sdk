# ups-sdk

English | [中文](./README.zh-CN.md)

Unofficial Go SDK for [UPS REST APIs](https://developer.ups.com/), generated from the official OpenAPI specs and wrapped with OAuth, retries, and a stable facade.

> **Not affiliated with UPS.** You need your own UPS Developer Portal account and OAuth credentials. See [DISCLAIMER.md](./DISCLAIMER.md).

## Why this layout

| Path | Purpose |
|------|---------|
| `ups/` | Hand-written facade (OAuth cache/refresh, Rate / Ship / Void / Track / Address Validation, errors) |
| `internal/generated/` | `oapi-codegen` output from UPS OpenAPI (committed so `go get` works) |
| `scripts/` + `generate.sh` | Fetch official YAML → patch → generate |
| `.cache/specs/` | Downloaded UPS YAML (**gitignored**, never commit) |

Official OpenAPI files are **not** stored in git. CI / `make generate` pulls them from [UPS-API/api-documentation](https://github.com/UPS-API/api-documentation).

## Install

```bash
go get github.com/befship/ups-sdk@latest
```

## Quick start

Credentials are passed in via `ups.Config` (plain strings). Reading them from the environment is optional and only a convenience in examples.

```go
package main

import (
	"context"
	"fmt"

	"github.com/befship/ups-sdk/ups"
)

func main() {
	client := ups.New(ups.Config{
		ClientID:     "your-client-id",     // required; pass from your secret store / DB / flags
		ClientSecret: "your-client-secret", // required
		AccountNumber: "your-account",     // optional; x-merchant-id on token
		Environment:  ups.EnvSandbox,      // or ups.EnvProduction
	})

	ctx := context.Background()
	tok, err := client.AccessToken(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("token ok, len=", len(tok))
}
```

### Rate (multi-package)

```go
resp, err := client.Rate(ctx, ups.RateRequest{
	RateRequest: &ups.RATERequestWrapper{ /* ... */ }.RateRequest,
}, &ups.RateOptions{RequestOption: "Shop"})
```

Build the body with the re-exported UPS schema types (`ups.RATERequestWrapper`, package arrays, etc.). See `examples/`.

### Create shipment / void / track

```go
shipResp, err := client.CreateShipment(ctx, body, nil)
voidResp, err := client.VoidShipment(ctx, shipmentID, nil)
trackResp, err := client.Track(ctx, "1Z...", nil)
```

### Full generated clients

For APIs beyond the facade helpers, use the typed clients (still via `ups`, without importing `internal/`):

```go
rc, err := client.RatingClient()
sc, err := client.ShippingClient()
tc, err := client.TrackingClient()
```

After `make generate`, every UPS API in `scripts/manifest.yaml` is available under `internal/generated/<package>`. Core helpers cover Rating, Shipping (ship/void/label recovery), Tracking, Address Validation, and OAuth.

## Regenerate from UPS specs

```bash
# Requires: go, python3 + PyYAML, network (or SPEC_LOCAL_DIR)
make generate

# Seed from a local clone of UPS-API/api-documentation
SPEC_LOCAL_DIR=/path/to/api-documentation make generate
```

Pipeline:

1. Download YAML from UPS GitHub (or copy from `SPEC_LOCAL_DIR`)
2. `scripts/preprocess.py` fixes known UPS OAS quirks
3. `oapi-codegen` writes Go under `internal/generated/`
4. Emits `ups/servers.gen.go` for per-API CIE/Production base URLs

## GitHub Actions

| Workflow | Trigger |
|----------|---------|
| `CI` | push / PR / manual |
| `Generate UPS SDK` | **manual only** (`workflow_dispatch`) — fetches latest specs, regenerates, opens a PR |

No cron / schedule — regenerate when you choose.

## License

- Hand-written code, scripts, examples, tests: **MIT** (`LICENSE`)
- Generated models/clients: derived from UPS OpenAPI; subject to UPS API Access Agreement (`DISCLAIMER.md`)

## Supported APIs (manifest)

Address Validation, Commerce Guard, Customs Detail, Dangerous Goods, Delivery Defense/Intercept, Forwarding, Global Checkout, Landed Cost, Locator, OAuth (client credentials + auth code), Paperless, Pickup (+ notifications / points), Pre-Notification, Protected Delivery Token, Quantum View, **Rating**, SCS Inventory/Logistics/Transportation, **Shipping**, Smart Pickup, Time in Transit, Trade Direct, Export Assure, Track Alert (+ Enhanced), **Tracking**, WorldEase, and more — see `scripts/manifest.yaml`.
