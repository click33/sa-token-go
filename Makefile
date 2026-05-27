MODULES := $(shell go work edit -json | python3 -c "import sys,json; print(' '.join(m['DiskPath'] for m in json.load(sys.stdin).get('Use',[])))")
LINT_MODULES := $(shell go work edit -json | python3 -c "import sys,json; print(' '.join(m['DiskPath'] for m in json.load(sys.stdin).get('Use',[]) if not m['DiskPath'].startswith('./examples') and not m['DiskPath'].startswith('./demo')))")

.PHONY: lint test test-race coverage build fmt vet

lint:
	@for dir in $(LINT_MODULES); do \
		echo "=== lint $$dir ==="; \
		(cd "$$dir" && golangci-lint run ./...) || exit 1; \
	done

test:
	@for dir in $(MODULES); do \
		echo "=== test $$dir ==="; \
		(cd "$$dir" && go test ./...) || exit 1; \
	done

test-race:
	@for dir in $(MODULES); do \
		echo "=== test-race $$dir ==="; \
		(cd "$$dir" && go test -race ./...) || exit 1; \
	done

coverage:
	@echo "mode: atomic" > coverage.out
	@for dir in $(MODULES); do \
		echo "=== coverage $$dir ==="; \
		(cd "$$dir" && go test -coverprofile=cover.tmp -covermode=atomic ./...) || exit 1; \
		tail -n +2 "$$dir/cover.tmp" >> coverage.out; \
		rm -f "$$dir/cover.tmp"; \
	done
	go tool cover -html=coverage.out -o coverage.html

build:
	@for dir in $(MODULES); do \
		echo "=== build $$dir ==="; \
		(cd "$$dir" && go build ./...) || exit 1; \
	done

fmt:
	gofmt -w .

vet:
	@for dir in $(MODULES); do \
		echo "=== vet $$dir ==="; \
		(cd "$$dir" && go vet ./...) || exit 1; \
	done
