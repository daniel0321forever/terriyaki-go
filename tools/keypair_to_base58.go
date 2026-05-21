package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mr-tron/base58/base58"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run keypair_to_base58.go /path/to/keypair.json")
		os.Exit(2)
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read file:", err)
		os.Exit(1)
	}

	var ints []int
	if err := json.Unmarshal(data, &ints); err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse keypair JSON (expected array of bytes):", err)
		os.Exit(1)
	}

	b := make([]byte, len(ints))
	for i, v := range ints {
		b[i] = byte(v)
	}

	// Print base58-encoded private key (this is what solana-go's PrivateKeyFromBase58 expects)
	s := base58.Encode(b)
	fmt.Println(s)
}
