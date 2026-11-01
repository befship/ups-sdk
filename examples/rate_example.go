//go:build ignore

// Rate example — pass credentials via ups.Config (env vars below are only for local demos).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/befship/ups-sdk/ups"
)

func main() {
	// In real apps, pass ClientID/ClientSecret from your config/secret store — not os.Getenv.
	client := ups.New(ups.Config{
		ClientID:       os.Getenv("UPS_CLIENT_ID"),
		ClientSecret:   os.Getenv("UPS_CLIENT_SECRET"),
		AccountNumber:  os.Getenv("UPS_ACCOUNT_NUMBER"),
		Environment:    envOr(ups.EnvSandbox),
		TransactionSrc: "ups-sdk-example",
	})

	shipperNum := os.Getenv("UPS_ACCOUNT_NUMBER")
	raw := fmt.Sprintf(`{
	  "RateRequest": {
	    "Request": {"TransactionReference": {"CustomerContext": "ups-sdk-example"}},
	    "Shipment": {
	      "Shipper": {
	        "Name": "Shipper",
	        "ShipperNumber": %q,
	        "Address": {
	          "AddressLine": ["123 Main St"],
	          "City": "Alpharetta",
	          "StateProvinceCode": "GA",
	          "PostalCode": "30005",
	          "CountryCode": "US"
	        }
	      },
	      "ShipTo": {
	        "Name": "Consignee",
	        "Address": {
	          "AddressLine": ["480 Park Ave"],
	          "City": "New York",
	          "StateProvinceCode": "NY",
	          "PostalCode": "10022",
	          "CountryCode": "US"
	        }
	      },
	      "ShipFrom": {
	        "Name": "Shipper",
	        "Address": {
	          "AddressLine": ["123 Main St"],
	          "City": "Alpharetta",
	          "StateProvinceCode": "GA",
	          "PostalCode": "30005",
	          "CountryCode": "US"
	        }
	      },
	      "Service": {"Code": "03", "Description": "Ground"},
	      "Package": [{
	        "PackagingType": {"Code": "02", "Description": "Package"},
	        "Dimensions": {
	          "UnitOfMeasurement": {"Code": "IN", "Description": "Inches"},
	          "Length": "10", "Width": "10", "Height": "10"
	        },
	        "PackageWeight": {
	          "UnitOfMeasurement": {"Code": "LBS", "Description": "Pounds"},
	          "Weight": "5"
	        }
	      }]
	    }
	  }
	}`, shipperNum)

	var req ups.RateRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		panic(err)
	}

	resp, err := client.Rate(context.Background(), req, &ups.RateOptions{RequestOption: "Rate"})
	if err != nil {
		panic(err)
	}
	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
}

func envOr(def string) string {
	if v := os.Getenv("UPS_ENV"); v != "" {
		return v
	}
	return def
}
