.PHONY: test test-unit test-repo-integration test-e2e test-blockchain-validation test-coverage lint build run clean ci-local

# Run all tests
test: test-unit test-repo-integration test-e2e

# Run the local CI sequence end-to-end (excludes Solana tests which require validator)
ci-local: test-unit test-repo-integration test-e2e validate-openapi

# Run unit tests only
test-unit:
	go test -v -race -coverprofile=coverage.out ./internal/domain/... ./internal/application/...

# Run repository integration tests (Postgres adapter)
test-repo-integration:
	@if ! docker info >/dev/null 2>&1; then \
		echo "Docker daemon is not available. Start Docker Desktop (or Colima/OrbStack) and retry 'make test-repo-integration'."; \
		exit 1; \
	fi
	go test -v -race -tags=integration ./internal/infrastructure/db/postgres/...

# Run end-to-end integration tests (HTTP -> service -> db) - excludes blockchain validation tests
test-e2e:
	go test -v -race ./tests/integration/...

# Run blockchain validation tests (requires: solana-cli, validator running on http://127.0.0.1:8899)
test-blockchain-validation:
	bash -c "source scripts/setup_solana_env.sh && \
		solana airdrop 2 \$$SOLANA_ORACLE_PUBKEY --url http://127.0.0.1:8899 && \
		go test -v -run TestSolanaE2E ./tests/blockchain_validation/..."

# Generate coverage report
test-coverage: test-unit
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Build binary
build:
	go build -o habitat-api ./internal/cmd/api_server

# Run development server
run:
	go run ./internal/cmd/api_server/main.go

# Clean build artifacts
clean:
	rm -f habitat-api coverage.out coverage.html

# Validate OpenAPI spec
validate-openapi:
	npx --yes @openapitools/openapi-generator-cli validate -i openapi.yaml

# Generate API documentation
docs:
	swag init -g internal/cmd/api_server/main.go
