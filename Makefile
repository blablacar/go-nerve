# Makefile for the Go Nerve Project
#
# To build the project, you need the gom package from https://github.com/mattn/gom
#

all: clean utest build

build:
	godep go build -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%H:%M:%S_UTC'` -X main.Version=`cat VERSION.txt`-`git rev-parse HEAD`" -o nerve

clean:
	rm -f nerve

utest:
	godep go test ./...

install:
	cp nerve ${GOPATH}/bin/nerve
