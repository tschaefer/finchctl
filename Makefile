Version := $(shell git describe --tags --dirty 2> /dev/null)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X github.com/tschaefer/finchctl/internal/version.Version=$(Version) -X github.com/tschaefer/finchctl/internal/version.GitCommit=$(GitCommit)"

.PHONY: all
all: fmt lint test dist

.PHONY: fmt
fmt:
	test -z $(shell gofmt -l .) || (echo "[WARN] Fix format issues" && exit 1)

.PHONY: lint
lint:
	test -z $(shell golangci-lint run >/dev/null || echo 1) || (echo "[WARN] Fix lint issues" && exit 1)

.PHONY: dist
dist:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/finchctl-linux-amd64 -ldflags $(LDFLAGS) .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/finchctl-linux-arm64 -ldflags $(LDFLAGS) .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/finchctl-darwin-amd64 -ldflags $(LDFLAGS) .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/finchctl-darwin-arm64 -ldflags $(LDFLAGS) .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/finchctl-windows-amd64 -ldflags $(LDFLAGS) .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o bin/finchctl-windows-arm64 -ldflags $(LDFLAGS) .

.PHONY: checksum
checksum:
	cd bin && \
	for f in finchctl-linux-amd64 finchctl-linux-arm64 \
		finchctl-darwin-amd64 finchctl-darwin-arm64 \
		finchctl-windows-amd64 finchctl-windows-arm64; do \
		sha256sum $$f > $$f.sha256; \
	done && \
	cd ..

.PHONY: test
test:
	test -z $(shell go test -v ./... 2>&1 >/dev/null || echo 1) || (echo "[WARN] Fix test issues" && exit 1)

.PHONY: clean
clean:
	rm -rf bin
