set quiet

_default: _help

_help:
    just --list

# Format go files
fmt:
    goimports -w .

# Lint go files
lint:
    golangci-lint run -v

# Run tests
test:
    go test -v -failfast -race -timeout 10m ./...

# Download go.mod dependencies
download:
    echo "Download go.mod dependencies"
    go mod download

# Check for vulnerabilities
vulncheck:
    govulncheck ./...

# Generate latest SDK
generate:
    go-sdk-gen generate --mod github.com/sumup/sumup-go --pkg sumup --name SumUp ./openapi.json
    gomarkdoc --output DOCUMENTATION.md ./...

# Install development dependencies
install-tools:
    command -v go-sdk-gen >/dev/null 2>&1 || go install github.com/sumup/go-sdk-gen@latest
    command -v gomarkdoc >/dev/null 2>&1 || go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
    command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
