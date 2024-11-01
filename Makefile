GO ?= go
GOBIN ?= $$($(GO) env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.60.3
TEST_COVERAGE ?= $(GOBIN)/go-test-coverage

.PHONY: install-golangcilint
install-golangcilint:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

# Runs lint on entire repo
.PHONY: lint
lint: install-golangcilint
	$(GOLANGCI_LINT) run ./...

# Runs tests on entire repo
.PHONY: test
test: 
	go test -timeout=3s -race -count=10 -failfast -shuffle=on -short ./... -coverprofile=./cover.short.profile -covermode=atomic -coverpkg=./...
	go test -timeout=10s -race -count=1 -failfast  -shuffle=on ./... -coverprofile=./cover.long.profile -covermode=atomic -coverpkg=./...


# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@latest


# Runs test coverage check
.PHONY: check-coverage
check-coverage: test
check-coverage: install-go-test-coverage
	$(TEST_COVERAGE) -config=./.testcoverage.yml

# View coverage profile
.PHONY: view-coverage
view-coverage: test
	go tool cover -html=cover.out -o=cover.html
	xdg-open cover.html