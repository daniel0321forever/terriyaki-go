#!/usr/bin/env bash
# Helper to export Solana-related env vars needed by the backend.
# Usage: source scripts/setup_solana_env.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

KEYPAIR_PATH_DEFAULT="$HOME/.config/solana/id.json"
KEYPAIR_PATH="${1:-$KEYPAIR_PATH_DEFAULT}"

if ! command -v solana-keygen >/dev/null 2>&1; then
  echo "warning: solana-keygen not found in PATH; SOLANA_ORACLE_PUBKEY won't be derived"
fi

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is required to run keypair_to_base58.go" >&2
  return 1
fi

if [ ! -f "$KEYPAIR_PATH" ]; then
  echo "error: keypair file not found: $KEYPAIR_PATH" >&2
  return 2
fi

echo "Using keypair: $KEYPAIR_PATH"

export SOLANA_RPC_ENDPOINT="http://127.0.0.1:8899"
export SOLANA_PROGRAM_ID="BgNjXioQqVNNihH4QCtjthDKAynZLVDixArQgmY7oRM4"

# Locate the keypair_to_base58.go tool in likely locations.
KEYPAIR_TOOL=""
if [ -f "$SCRIPT_DIR/../tools/keypair_to_base58.go" ]; then
  KEYPAIR_TOOL="$SCRIPT_DIR/../tools/keypair_to_base58.go"
elif [ -f "$SCRIPT_DIR/../backend/tools/keypair_to_base58.go" ]; then
  KEYPAIR_TOOL="$SCRIPT_DIR/../backend/tools/keypair_to_base58.go"
elif [ -f "$SCRIPT_DIR/../../backend/tools/keypair_to_base58.go" ]; then
  KEYPAIR_TOOL="$SCRIPT_DIR/../../backend/tools/keypair_to_base58.go"
fi

if [ -z "$KEYPAIR_TOOL" ]; then
  echo "error: could not find backend/tools/keypair_to_base58.go; please run from repo root or backend folder" >&2
  return 3
fi

echo "Generating base58 private key from keypair (this may take a moment)..."
export SOLANA_ORACLE_PRIVATE_KEY="$(go run "$KEYPAIR_TOOL" "$KEYPAIR_PATH")"

if command -v solana-keygen >/dev/null 2>&1; then
  export SOLANA_ORACLE_PUBKEY="$(solana-keygen pubkey "$KEYPAIR_PATH")"
else
  echo "solana-keygen not available; please set SOLANA_ORACLE_PUBKEY manually"
fi

echo "Exported SOLANA_RPC_ENDPOINT=$SOLANA_RPC_ENDPOINT"
echo "Exported SOLANA_PROGRAM_ID=$SOLANA_PROGRAM_ID"
echo "Exported SOLANA_ORACLE_PUBKEY=${SOLANA_ORACLE_PUBKEY:-<not-set>}"
echo "SOLANA_ORACLE_PRIVATE_KEY is set (hidden)"

echo "Done. Start the backend in the same shell, for example:"
echo "  cd backend && go run ./internal/cmd/api_server/"
