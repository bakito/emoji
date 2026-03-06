# Include toolbox tasks
include ./.toolbox.mk

# Run go golanci-lint
lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test:
	go test ./... -cover -coverprofile=coverage.out
