# Disclaimer

## Not affiliated with UPS

This project is **not** affiliated with, endorsed by, or sponsored by United Parcel Service of America, Inc. (“UPS”) or any of its affiliates. “UPS” and related marks are trademarks of their respective owners.

## UPS API Access Agreement

UPS REST APIs are subject to the [UPS API Access Agreement](https://www.ups.com/upsdeveloperkit/assets/html/serviceAgreement.html) and UPS Developer Portal terms.

- UPS grants a **limited, revocable, non-transferable** license to download official OpenAPI specifications and to generate clients for integrating with UPS APIs.
- **Do not** redistribute UPS official OpenAPI YAML/JSON as the primary content of a public repository. This project fetches specs at generate/CI time and does **not** commit them.
- Use of this SDK for commercial purposes requires your own compliance with UPS terms, including obtaining your own developer credentials.
- UPS terms generally prohibit using the APIs to build an independent API gateway or to resell UPS API access as a standalone product.

## Generated code

Files under `internal/generated/` are auto-generated from UPS public OpenAPI specifications published at [UPS-API/api-documentation](https://github.com/UPS-API/api-documentation). Derived model/client code follows UPS documentation licensing; hand-written wrappers, scripts, and examples are MIT-licensed (see `LICENSE`).

## Credentials

This SDK does **not** ship UPS secrets. You must register at the UPS Developer Portal and obtain Sandbox/Production OAuth client credentials yourself.

## No warranty

The software is provided “as is”, without warranty of any kind. API behavior, schemas, and endpoints may change when UPS updates their specifications.
