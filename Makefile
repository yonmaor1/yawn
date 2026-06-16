BINARY := yawn
DIST   := dist
PREFIX ?= /usr/local

.PHONY: build test install clean

build:
	bin/go build -o $(DIST)/$(BINARY) .

test:
	bin/go test ./...

install: build
	install -d $(PREFIX)/bin
	install -m 0755 $(DIST)/$(BINARY) $(PREFIX)/bin/$(BINARY)

clean:
	rm -rf $(DIST)
