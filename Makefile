all: tracker-ui-backend

VERSION=`git rev-parse --short HEAD`
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
OS_NAME := $(shell uname -s | tr A-Z a-z)
SED_COMMAND:=sed

ifeq ($(OS_NAME),darwin)
	SED_COMMAND=gsed
endif

tracker-ui-backend:
	make -C cmd tracker-ui-backend
