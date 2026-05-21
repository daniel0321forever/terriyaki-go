# Blockchain Validation Tests

This directory contains special validation tests for the **Solana payment adapter** and on-chain settlement flows. These tests require external infrastructure and tools that are not available in standard CI environments.

## Prerequisites

Before running these tests, ensure the following are available on your machine:

1. **Solana Validator** running on `http://127.0.0.1:8899`
   - Start with: `solana-test-validator --reset`
   - Keep it running in a separate terminal during test execution.

2. **Solana CLI Tools** (v1.18+)
   - Required: `solana-keygen`, `solana` command
   - Install: https://docs.solana.com/cli/install-solana-cli-tools

3. **Backend Running** with Solana environment variables exported
   - Source the setup script: `source backend/scripts/setup_solana_env.sh`
   - Then start the backend: `cd backend && go run ./internal/cmd/api_server/`

4. **Local Solana Program Deployed**
   - Build and deploy: `cd solana && anchor build && anchor deploy --provider.cluster localnet`
   - Verify deployment: `solana program show BgNjXioQqVNNihH4QCtjthDKAynZLVDixArQgmY7oRM4 --url http://127.0.0.1:8899`

## Running the Tests

### Option 1: Using the Makefile (Recommended)

```bash
# From backend/ directory, run all Solana tests
make test-solana

# Or with additional logging
make test-solana 2>&1 | tee solana-test.log
```

### Option 2: Using Go directly

```bash
# From backend/ directory
go test -v -race ./tests/blockchain_validation/...

# Or run a specific test
go test -v -run TestSolanaE2E ./tests/blockchain_validation/...
```

## What These Tests Validate

1. **Payment Method Onboarding** — Adding a Solana wallet as a payment method.
2. **Adapter Integration** — The Solana payment gateway adapter building and functioning.
3. **On-Chain Interaction** — Pledge creation, resolution, and confirmation against the local validator.
4. **Idempotency & Retry Safety** — Settlement operations handle retries correctly.

## CI/CD Integration

**These tests are intentionally excluded from the default CI/CD pipeline** (`make ci-local` and `make test-e2e`) because:
- They require a running Solana validator (not available in CI containers).
- They depend on local Solana CLI tools that may not be installed.
- They add significant latency to the test suite.

To include them in a CI pipeline, you would need to:
1. Spin up a Solana validator as a service container (similar to Postgres/Redis in CI).
2. Ensure Solana CLI tools are installed in the CI image.
3. Deploy the Anchor program before running tests.

For now, run these tests **locally only** as part of the Solana adapter development workflow.

## Troubleshooting

### Test Fails with "skipped" Status
- **Cause:** Solana CLI tools not in PATH or keypair file missing.
- **Fix:** Install Solana CLI and ensure `solana-keygen pubkey` works.

### Test Fails with "Internal Server Error" (500)
- **Cause:** Backend panicked or error not properly handled in the payment method endpoint.
- **Fix:** Check backend logs for panics. Run `go test ./tests/integration/...` first to ensure basic E2E tests pass.

### Transaction Confirmation Timeout
- **Cause:** Local validator is slow or not synced.
- **Fix:** Restart the validator and ensure RPC endpoint responds: `curl http://127.0.0.1:8899 -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","id":1,"method":"getHealth"}'`

## References

- [Solana CLI Docs](https://docs.solana.com/cli)
- [Anchor Framework Docs](https://www.anchor-lang.com)
- [Solana RPC API](https://docs.solana.com/api)
