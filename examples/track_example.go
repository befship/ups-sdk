//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/befship/ups-sdk/ups"
)

func main() {
	inquiry := os.Getenv("UPS_TRACKING_NUMBER")
	if inquiry == "" {
		fmt.Fprintln(os.Stderr, "set UPS_TRACKING_NUMBER")
		os.Exit(1)
	}
	client := ups.New(ups.Config{
		ClientID:     os.Getenv("UPS_CLIENT_ID"),
		ClientSecret: os.Getenv("UPS_CLIENT_SECRET"),
		Environment:  envOr(ups.EnvSandbox),
	})
	resp, err := client.Track(context.Background(), inquiry, nil)
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
