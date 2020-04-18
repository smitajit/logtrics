SRC_FILES = $(shell go list -f '{{range .GoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}' ./...)
TEST_FILES = $(shell go list -f '{{range .TestGoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}' ./...)
APP_BIN = logtrics
TAG ?= $(shell TZ=UTC date +%Y%m%d_%H%M%S)
PREFIX ?= /usr

all: build

build: $(APP_BIN)
install: $(PREFIX)/bin/$(APP_BIN)

.PHONY: clean
clean:
	rm -rf $(APP_BIN) coverage.out

$(APP_BIN): $(SRC_FILES) ./cmd/main.go
	@go build  -ldflags '-w -s -X main.BuildDate=$(shell date +%F)' -o $@ ./cmd/main.go

$(PREFIX)/bin/$(APP_BIN): $(APP_BIN)
	install -p -D -m 0755 $< $@


coverage.out: $(TEST_FILES) $(SRC_FILES)
	@go test -v -cover -coverprofile $(@) ./...

.PHONY: cover
cover: coverage.out
	@go tool cover -func $<

.PHONY: vet
vet:
	@go vet ./...

.PHONY: fmt
fmt:
	@bash -c "diff -u <(echo -n) <(gofmt -d ./)"

.PHONY: test
test: cover vet fmt
