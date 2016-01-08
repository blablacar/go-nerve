# Makefile for the Go Nerve Project
#
# To build the project, you need the gom package from https://github.com/mattn/gom
#

GOMCMD=gom

all: clean dep-install build

dep-install:
	$(GOMCMD) install

build:
	$(GOMCMD) build nerve/nerve
	mv nerve bin/.

clean:
	rm -f bin/*
	rm -rf _vendor

install:
	cp bin/nerve /usr/local/bin/nerve
