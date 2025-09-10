.PHONY: build install clean test

# Binary name
BINARY=ccl

# Installation directory
INSTALL_DIR=$(HOME)/.local/bin

build:
# go build -ldflags "-w -s" microservice.go
	go build -ldflags "-w -s" -o $(BINARY) main.go

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/
	chmod +x $(INSTALL_DIR)/$(BINARY)
	@echo "CCL installed to $(INSTALL_DIR)/$(BINARY)"

clean:
	rm -f $(BINARY)

test:
	go test ./...

all: clean build