build:
    go build ./...

test:
    go run gotest.tools/gotestsum@latest --format testname ./...

lint *args:
    golangci-lint run -D exhaustruct -D err113 -D godox {{ args }}

check-fmt:
    ./hack/error-on-diff.sh just fmt

fmt:
    go mod tidy
    treefmt
