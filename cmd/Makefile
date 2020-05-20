all: tracker-controller

VERSION=`git rev-parse --short HEAD`
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
ARCH := $(shell uname -m)
OS_NAME := $(shell uname -s | tr A-Z a-z)

ifeq ($(ARCH),aarch64)
	ARCH=arm64
endif

ifeq ($(ARCH),x86_64)
	ARCH=amd64
endif

prebuild:
	mkdir -p bin/arm64/linux bin/amd64/linux bin/amd64/windows bin/amd64/darwin

tracker-controller: prebuild
	GO111MODULE=on GOOS=$(OS_NAME) GOARCH=$(ARCH) go build -o bin/$(ARCH)/$(OS_NAME)/tracker-controller tracker-controller/*.go 
