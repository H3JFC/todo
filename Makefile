BINARY             := todo
GOBIN              := $(shell go env GOPATH)/bin
BIN_TARGET         := $(HOME)/bin/$(BINARY)
COMPLETION         := $(HOME)/.oh-my-zsh/completions/_todo
COMPLETION_BACKUP  := $(HOME)/.settings/other/_todo
ZCOMPDUMP          := $(HOME)/.zcompdump

.PHONY: all test install go-install completion zcompdump clean

all: install

test:
	go test ./...

install: clean go-install completion zcompdump

go-install:
	go install .
	@mkdir -p $(dir $(BIN_TARGET))
	cp $(GOBIN)/$(BINARY) $(BIN_TARGET)

completion:
	@mkdir -p $(dir $(COMPLETION))
	go run . completion zsh > $(COMPLETION)
	@mkdir -p $(dir $(COMPLETION_BACKUP))
	cp $(COMPLETION) $(COMPLETION_BACKUP)

zcompdump:
	rm -f $(ZCOMPDUMP)*

clean:
	rm -f $(COMPLETION) $(COMPLETION_BACKUP) $(ZCOMPDUMP)*
