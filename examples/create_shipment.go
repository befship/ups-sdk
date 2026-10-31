//go:build ignore

// Create shipment example — set UPS_CLIENT_ID, UPS_CLIENT_SECRET, UPS_ACCOUNT_NUMBER.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/befship/ups-sdk/ups"
)

func main() {
	client := ups.New(ups.Config{
		ClientID:       os.Getenv("UPS_CLIENT_ID"),
		ClientSecret:   os.Getenv("UPS_CLIENT_SECRET"),
		AccountNumber:  os.Getenv("UPS_ACCOUNT_NUMBER"),
		Environment:    envOr(ups.EnvSandbox),
		TransactionSrc: "ups-sdk-example",
	})

	acct := os.Getenv("UPS_ACCOUNT_NUMBER")
	raw := fmt.Sprintf(`{
	  "ShipmentRequest": {
	    "Request": {"RequestOption": "nonvalidate", "TransactionReference": {"CustomerContext": "ups-sdk-example"}},
	    "Shipment": {
	      "Description": "Ship SDK example",
	      "Shipper": {
	        "Name": "Shipper",
	        "AttentionName": "Shipper",
	        "ShipperNumber": %q,
	        "Phone": {"Number": "5555555555"},
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
	        "AttentionName": "Consignee",
	        "Phone": {"Number": "5555555555"},
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
	        "AttentionName": "Shipper",
	        "Phone": {"Number": "5555555555"},
	        "Address": {
	          "AddressLine": ["123 Main St"],
	          "City": "Alpharetta",
	          "StateProvinceCode": "GA",
	          "PostalCode": "30005",
	          "CountryCode": "US"
	        }
	      },
	      "PaymentInformation": {
	        "ShipmentCharge": {
	          "Type": "01",
	          "BillShipper": {"AccountNumber": %q}
	        }
	      },
	      "Service": {"Code": "03", "Description": "Ground"},
	      "Package": [{
	        "Description": "Package",
	        "Packaging": {"Code": "02", "Description": "Package"},
	        "Dimensions": {
	          "UnitOfMeasurement": {"Code": "IN", "Description": "Inches"},
	          "Length": "10", "Width": "10", "Height": "10"
	        },
	        "PackageWeight": {
	          "UnitOfMeasurement": {"Code": "LBS", "Description": "Pounds"},
	          "Weight": "5"
	        }
	      }]
	    },
	    "LabelSpecification": {
	      "LabelImageFormat": {"Code": "GIF", "Description": "GIF"},
	      "HTTPUserAgent": "ups-sdk"
	    }
	  }
	}`, acct, acct)

	var req ups.ShipmentRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		panic(err)
	}

	resp, err := client.CreateShipment(context.Background(), req, nil)
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
