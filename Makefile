BINARY := tv
GO := go
NG := npx ng

.PHONY: build web test clean build-all

build: web
	$(GO) build -o $(BINARY).exe ./cmd/$(BINARY)/

web:
	cd web && $(NG) build

test:
	$(GO) test -v ./...

build-all:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY)-windows-amd64.exe ./cmd/$(BINARY)/
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY)-linux-amd64 ./cmd/$(BINARY)/
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY)-darwin-amd64 ./cmd/$(BINARY)/

clean:
	del /f /q $(BINARY).exe $(BINARY)-windows-*.exe $(BINARY)-linux-* $(BINARY)-darwin-* 2>nul || true
