include $(GOROOT)/src/Make.inc

all: make

make:
	$(MAKE) -C src all

testpackage:
	$(MAKE) -C src testpackage

clean:
	$(MAKE) -C src clean

nuke:
	$(MAKE) -C src nuke

install:
	$(MAKE) -C src install

test:
	$(MAKE) -C src test

bench:
	$(MAKE) -C src bench

prepare-genmake:
	$(MAKE) -C src/cmd/genmake nuke install
	genmake

prepare:
	$(MAKE) prepare-genmake
	$(MAKE) gofmt

gofmt:
	$(MAKE) -C src gofmt

govet:
	$(MAKE) -C src govet

.PHONY: all testpackage clean nuke install test bench prepare prepare-genmake prepare-genschema gofmt govet
