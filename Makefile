# Color
GREEN := \033[0;32m
NC := \033[0m

.PHONY: test
test: fmt
	@echo "$(GREEN)» test$(NC)"
	go test ./...

.PHONY: bench
bench:
	@echo "$(GREEN)» bench$(NC)"
	go test -bench=. ./...

.PHONY: fmt
fmt:
	@echo "$(GREEN)» fmt$(NC)"
	go fmt ./...

.PHONY: build
build: test
	@echo "$(GREEN)» build$(NC)"
	go build -o fixture/sqlsplitter cmd/main.go

.PHONY: cover
cover: test
	@echo "$(GREEN)» check test code coverage$(NC)"
	@go test $(shell go list -f '{{if ne .Name "main"}}{{.ImportPath}}{{end}}' ./...) -coverprofile=fixture/coverage.data
	@go tool cover -html=fixture/coverage.data -o fixture/coverage.html
	@go tool cover -func=fixture/coverage.data -o fixture/coverage.txt
