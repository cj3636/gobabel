.PHONY: test bench run fmt vet
run:
	go run ./cmd/gobabel serve

test:
	go test ./...

bench:
	go test -bench=. ./...

fmt:
	gofmt -w .

vet:
	go vet ./...
