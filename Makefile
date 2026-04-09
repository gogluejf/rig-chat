BINARY := bin/rig-chat

.PHONY: build install clean

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) .

install: build
	cp $(BINARY) /usr/local/bin/rig-chat 2>/dev/null || \
		cp $(BINARY) $(HOME)/go/bin/rig-chat

clean:
	rm -f $(BINARY)
