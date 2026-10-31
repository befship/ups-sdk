// Package ups is an unofficial Go client for UPS REST APIs.
//
// # Overview
//
//   - OAuth2 client-credentials with in-memory token cache
//   - Sandbox (CIE) vs Production base URLs per API
//   - Convenience methods: Rate, CreateShipment, VoidShipment, LabelRecovery, Track, ValidateAddress
//   - Accessors for every UPS API listed in scripts/manifest.yaml (*Client methods)
//
// Generated OpenAPI clients live in internal/generated and are not imported by
// downstream modules directly; use this package instead.
//
// This project is not affiliated with UPS. See DISCLAIMER.md at the module root.
package ups
