# Make this makefile self-documented with target `help`
.PHONY: help
.DEFAULT_GOAL := help
help: ## Show help
	@grep -Eh '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: ## Format go files
	golangci-lint fmt --verbose

.PHONY: lint
lint: ## Lint go files
	golangci-lint run --verbose

.PHONY: lint-fix
lint-fix: ## Lint go files and apply auto-fixes
	golangci-lint run --verbose --fix

.PHONY: test
test: ## Run tests
	go test -v -failfast -race -timeout 10m ./...

.PHONY: download
download:
	@echo Download go.mod dependencies
	@go mod download

.PHONY: vulncheck
vulncheck: ## Check for Vulnerabilities (make sure you have the tools install: `make install-tools`)
	govulncheck ./...

.PHONY: vulncheck-sarif
vulncheck-sarif: ## Check for Vulnerabilities
	govulncheck -format=sarif ./... > govulncheck.sarif

.PHONY: generate
generate: ## Generate latest SDK
	cd codegen && go run ./... generate --out ../ ../openapi.json
	gomarkdoc --repository.url https://github.com/sumup/sumup-go --repository.default-branch main --exclude-dirs ./codegen --output DOCUMENTATION.md ./...

.PHONY: install-tools
install-tools: # Install development dependencies
	cd codegen && go install ./cmd/go-sdk-gen
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
