.PHONY: build test check fmt dashboard dashboard-check app report-assets
build:
	go build -o bin/konkour-bench ./cmd/konkour-bench
test:
	go test -race ./...
check:
	@test -z "$$(gofmt -l cmd internal)" || (gofmt -l cmd internal; exit 1)
	go vet ./...
	go test -race ./...
	go build ./...
	npx --yes oxlint@1.88.0 internal/report/report.js internal/report/charts.js internal/report/logos.js internal/report/report.test.cjs internal/report/build-math.mjs
	node --test internal/report/report.test.cjs
fmt:
	gofmt -w cmd internal
dashboard:
	npm --prefix dashboard ci --no-audit --no-fund
	npm --prefix dashboard run build
dashboard-check:
	npm --prefix dashboard run check
app: build dashboard
	./bin/konkour-bench serve
report-assets:
	node internal/report/build-math.mjs
