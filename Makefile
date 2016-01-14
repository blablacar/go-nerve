# Makefile for the Go Nerve Project
#
# To build the project, you need the gom package from https://github.com/mattn/gom
#

GOMCMD=gom

all: clean dep-install build

dep-install:
	$(GOMCMD) install

build:
	$(GOMCMD) build -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%H:%M:%S_UTC'` -X main.Version=`cat VERSION.txt`-`git rev-parse HEAD`" nerve/nerve
	mv nerve bin/.

clean:
	rm -f bin/*
	rm -rf _vendor

install:
	cp bin/nerve /usr/local/bin/nerve
