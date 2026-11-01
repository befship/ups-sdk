package rating

import (
	"encoding/json"
	"testing"
)

func TestNegotiatedRateChargesBaseServiceChargeIsObject(t *testing.T) {
	raw := []byte(`{
		"BaseServiceCharge": {"CurrencyCode":"EUR","MonetaryValue":"56.03"},
		"TotalCharge": {"CurrencyCode":"EUR","MonetaryValue":"87.42"}
	}`)
	var got RatedShipmentNegotiatedRateCharges
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.BaseServiceCharge == nil || got.BaseServiceCharge.MonetaryValue != "56.03" {
		t.Fatalf("BaseServiceCharge = %#v", got.BaseServiceCharge)
	}
}

func TestRATEResponseWrapperNegotiatedObject(t *testing.T) {
	raw := []byte(`{
	  "RateResponse": {
	    "Response": {"ResponseStatus": {"Code": "1", "Description": "Success"}},
	    "RatedShipment": [{
	      "Service": {"Code": "07"},
	      "BillingWeight": {"UnitOfMeasurement": {"Code": "KGS"}, "Weight": "1.0"},
	      "TransportationCharges": {"CurrencyCode": "EUR", "MonetaryValue": "88.30"},
	      "ServiceOptionsCharges": {"CurrencyCode": "EUR", "MonetaryValue": "0.00"},
	      "TotalCharges": {"CurrencyCode": "EUR", "MonetaryValue": "88.30"},
	      "NegotiatedRateCharges": {
	        "BaseServiceCharge": {"CurrencyCode": "EUR", "MonetaryValue": "56.03"},
	        "TotalCharge": {"CurrencyCode": "EUR", "MonetaryValue": "87.42"}
	      },
	      "RatedPackage": [{"Weight": "1.0"}]
	    }]
	  }
	}`)
	var wrap RATEResponseWrapper
	if err := json.Unmarshal(raw, &wrap); err != nil {
		t.Fatalf("RATEResponseWrapper: %v", err)
	}
	neg := wrap.RateResponse.RatedShipment[0].NegotiatedRateCharges
	if neg == nil || neg.BaseServiceCharge == nil || neg.BaseServiceCharge.MonetaryValue != "56.03" {
		t.Fatalf("NegotiatedRateCharges = %#v", neg)
	}
}
