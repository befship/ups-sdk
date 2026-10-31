.PHONY: generate generate-core fetch-specs test lint fmt tidy clean

# Full regenerate: fetch UPS OpenAPI specs, patch, generate all Go clients.
generate:
	bash generate.sh

# Core APIs only (faster local iteration).
generate-core:
	ONLY=oauthclientcredentials,rating,shipping,tracking,addressvalidation bash generate.sh

# Force re-download of official specs (still not committed).
fetch-specs:
	FORCE_FETCH=1 bash generate.sh

test:
	go test ./... -count=1

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

fmt:
	gofmt -w ups examples
	find internal/generated -name '*.go' -print0 2>/dev/null | xargs -0 gofmt -w

tidy:
	go mod tidy

clean:
	rm -rf .cache/specs
