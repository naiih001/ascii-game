.PHONY: game server fmt test tidy help

GO ?= go
ADDR ?= 127.0.0.1:7777
GAME_CMD := ./cmd/game
SERVER_CMD := ./cmd/server

help:
	@printf "Targets:\n"
	@printf "  make game    Run the multiplayer client\n"
	@printf "  make server  Run the multiplayer server\n"
	@printf "  make fmt     Format Go sources\n"
	@printf "  make test    Run Go tests\n"
	@printf "  make tidy    Sync Go module dependencies\n"

game:
	$(GO) run $(GAME_CMD) -addr $(ADDR)

server:
	$(GO) run $(SERVER_CMD) -addr $(ADDR)

fmt:
	$(GO) fmt ./...

test:
	GOCACHE=/tmp/go-build $(GO) test ./...

tidy:
	$(GO) mod tidy
