CYAN   := \033[0;36m
GREEN  := \033[0;32m
YELLOW := \033[0;33m
NC     := \033[0m

export GOWORK=off

make:
	@go run .

build:
	@go build .

check:
	@printf "$(CYAN)*** go fmt…$(NC)\n"
	@go fmt ./...
	@printf "$(CYAN)*** go update…$(NC)\n"
	@go get -u ./...
	@printf "$(CYAN)*** go vet…$(NC)\n"
	@go vet ./...
	@printf "$(CYAN)*** staticcheck…$(NC)\n"
	@staticcheck ./...
	@printf "$(CYAN)*** golangci-lint…$(NC)\n"
	@golangci-lint run ./...
	@printf "$(CYAN)*** govulncheck…$(NC)\n"
	@govulncheck ./...
	@printf "$(CYAN)*** gosec…$(NC)\n"
	@gosec -severity=HIGH -confidence=HIGH ./...
	@printf "$(CYAN)*** osv-scanner…$(NC)\n"
	@osv-scanner .
	@printf "$(GREEN)*** Lint & security checks passed! 🎉$(NC)\n"
