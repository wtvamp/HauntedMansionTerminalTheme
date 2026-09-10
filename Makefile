# Haunted Mansion terminal theme.
#
# The command names here are a contract documented in CLAUDE.md — if you change
# one, change it there too.

GO      ?= go
BIN     := bin
PREFIX  ?= $(HOME)/.local
ZSHDIR  ?= $(HOME)/.zshrc.d

# CONJURE_TARGET selects a single emulator; empty means all of them.
CONJURE_TARGET ?=


.DEFAULT_GOAL := help
.PHONY: help build generate install uninstall test check demo fmt clean

help: ## Show this help
	@grep -hE '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Compile conjure, doombuggy and portrait into bin/
	@$(GO) build -o $(BIN)/conjure ./cmd/conjure
	@$(GO) build -o $(BIN)/doombuggy ./cmd/doombuggy
	@$(GO) build -o $(BIN)/portrait ./cmd/portrait
	@echo "  built $(BIN)/conjure, $(BIN)/doombuggy and $(BIN)/portrait"

generate: build ## Render the palette into dist/
	@CONJURE_TARGET=$(CONJURE_TARGET) $(BIN)/conjure

test: ## Run the test suite
	@$(GO) test ./...

# check is what CI runs. It regenerates in memory and compares against the
# committed dist/, so a palette change that forgot to regenerate fails here
# rather than shipping a theme whose files disagree with its source.
check: build test ## Verify dist/ is not stale, then test
	@$(BIN)/conjure -check

fmt: ## Format Go sources
	@$(GO) fmt ./...

demo: build ## Run the ride without installing anything
	@$(BIN)/doombuggy

install: generate ## Install binaries to ~/.local/bin and shell files to ~/.zshrc.d
	@mkdir -p $(PREFIX)/bin $(ZSHDIR)
	@install -m 0755 $(BIN)/doombuggy $(PREFIX)/bin/doombuggy
	@install -m 0755 $(BIN)/conjure $(PREFIX)/bin/conjure
	@ln -sf $(CURDIR)/shell/greeting.sh $(ZSHDIR)/50-haunted-mansion-greeting.zsh
	@ln -sf $(CURDIR)/shell/haunted-mansion.zsh-theme $(ZSHDIR)/49-haunted-mansion-theme.zsh
	@echo "  installed. add this to ~/.zshrc if it is not there already:"
	@echo ""
	@echo "      export HM_ROOT=$(CURDIR)"
	@echo "      for f in $(ZSHDIR)/*.zsh(N); do source \$$f; done"
	@echo ""
	@echo "  then point your terminal at a file in dist/ — see README.md"

uninstall: ## Remove what install put in place
	@rm -f $(PREFIX)/bin/doombuggy $(PREFIX)/bin/conjure
	@rm -f $(ZSHDIR)/50-haunted-mansion-greeting.zsh $(ZSHDIR)/49-haunted-mansion-theme.zsh
	@echo "  removed. HM_ROOT and the source line in ~/.zshrc are yours to delete."

clean: ## Remove build output (not dist/, which is committed)
	@rm -rf $(BIN)
