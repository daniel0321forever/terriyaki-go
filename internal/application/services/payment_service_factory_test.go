package services

import (
	"os"
	"testing"

	solanaGo "github.com/gagliardetto/solana-go"
)

func TestLoadSolanaConfigFromEnv(t *testing.T) {
	// Backup env
	oldRPC := os.Getenv("SOLANA_RPC_ENDPOINT")
	oldProgID := os.Getenv("SOLANA_PROGRAM_ID")
	oldOraclePubkey := os.Getenv("SOLANA_ORACLE_PUBKEY")
	oldOraclePrivkey := os.Getenv("SOLANA_ORACLE_PRIVATE_KEY")

	defer func() {
		_ = os.Setenv("SOLANA_RPC_ENDPOINT", oldRPC)
		_ = os.Setenv("SOLANA_PROGRAM_ID", oldProgID)
		_ = os.Setenv("SOLANA_ORACLE_PUBKEY", oldOraclePubkey)
		_ = os.Setenv("SOLANA_ORACLE_PRIVATE_KEY", oldOraclePrivkey)
	}()

	// 1. Missing env vars
	_ = os.Unsetenv("SOLANA_RPC_ENDPOINT")
	_, _, _, _, err := LoadSolanaConfigFromEnv()
	if err == nil {
		t.Fatal("expected error when SOLANA_RPC_ENDPOINT is missing")
	}

	_ = os.Setenv("SOLANA_RPC_ENDPOINT", "http://localhost:8899")
	_ = os.Unsetenv("SOLANA_PROGRAM_ID")
	_, _, _, _, err = LoadSolanaConfigFromEnv()
	if err == nil {
		t.Fatal("expected error when SOLANA_PROGRAM_ID is missing")
	}

	// 2. Invalid keys
	_ = os.Setenv("SOLANA_PROGRAM_ID", "invalid-key")
	_ = os.Setenv("SOLANA_ORACLE_PUBKEY", "invalid-key")
	_ = os.Setenv("SOLANA_ORACLE_PRIVATE_KEY", "invalid-key")
	_, _, _, _, err = LoadSolanaConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid keys")
	}

	// 3. Valid config
	priv, err := solanaGo.NewRandomPrivateKey()
	if err != nil {
		t.Fatalf("failed to generate random private key: %v", err)
	}
	pub := priv.PublicKey()

	_ = os.Setenv("SOLANA_RPC_ENDPOINT", "http://localhost:8899")
	_ = os.Setenv("SOLANA_PROGRAM_ID", pub.String())
	_ = os.Setenv("SOLANA_ORACLE_PUBKEY", pub.String())
	_ = os.Setenv("SOLANA_ORACLE_PRIVATE_KEY", priv.String())

	rpc, prog, oracle, key, err := LoadSolanaConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error for valid config, got: %v", err)
	}

	if rpc != "http://localhost:8899" {
		t.Errorf("expected rpc http://localhost:8899, got %s", rpc)
	}
	if prog.String() != pub.String() {
		t.Errorf("expected program ID %s, got %s", pub.String(), prog.String())
	}
	if oracle.String() != pub.String() {
		t.Errorf("expected oracle pubkey %s, got %s", pub.String(), oracle.String())
	}
	if key.String() != priv.String() {
		t.Errorf("expected private key %s, got %s", priv.String(), key.String())
	}
}
