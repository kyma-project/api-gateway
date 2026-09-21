//go:build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/kyma-project/api-gateway/internal/signature"
)

func main() {
	encoded := os.Getenv("GARDENER_APIRULE_ACCESS")
	if encoded == "" {
		fmt.Fprintln(os.Stderr, "GARDENER_APIRULE_ACCESS is not set")
		os.Exit(1)
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Fprintf(os.Stderr, "base64 decode failed: %v\n", err)
		os.Exit(1)
	}

	msg, _, err := signature.DecryptAndVerifySignature(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "signature verification failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("signed message: %s\n", msg)
}
