BINARY := tv
GO := go

.PHONY: build test clean build-all

build:
	$(GO) build -o $(BINARY).exe ./cmd/$(BINARY)/

test:
	$(GO) test -v ./...

build-all:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY)-windows-amd64.exe ./cmd/$(BINARY)/
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY)-linux-amd64 ./cmd/$(BINARY)/
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY)-darwin-amd64 ./cmd/$(BINARY)/

clean:
	rm -f $(BINARY).exe $(BINARY)-windows-*.exe $(BINARY)-linux-* $(BINARY)-darwin-*
