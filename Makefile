.PHONY: build test check fmt dashboard dashboard-check app
build:
	go build -o bin/konkour-bench ./cmd/konkour-bench
test:
	go test -race ./...
check:
	@test -z "$$(gofmt -l cmd internal)" || (gofmt -l cmd internal; exit 1)
	go vet ./...
	go test -race ./...
	go build ./...
	npx --yes oxlint@1.88.0 internal/report/report.js
fmt:
	gofmt -w cmd internal
dashboard:
	npm --prefix dashboard ci --no-audit --no-fund
	npm --prefix dashboard run build
dashboard-check:
	npm --prefix dashboard run check
app: build dashboard
	./bin/konkour-bench serve
