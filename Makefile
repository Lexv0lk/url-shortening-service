COVERAGE_FILE ?= coverage.out

TARGET ?= urlshorteningservice # CHANGE THIS TO YOUR BINARY NAME

.PHONY: build
build:
	@echo "Processing go build for target ${TARGET}"
	@mkdir -p .bin
	@go build -o ./bin/${TARGET} ./cmd/${TARGET}

## test: run all tests
.PHONY: test
test:
	@go test -coverpkg='url-shortening-service/...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'
