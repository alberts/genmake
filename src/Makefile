all: make

make: pkg.install cmd.make
testpackage: pkg.testpackage cmd.testpackage
clean: pkg.nuke cmd.clean
nuke: pkg.nuke cmd.nuke
install: make
test: pkg.test cmd.test
bench: pkg.bench cmd.bench

%.install:
	$(MAKE) -C $* install

# compile all packages before any command
%.make: pkg.install
	$(MAKE) -C $*

pkg.testpackage: pkg.install
	$(MAKE) -C pkg testpackage

cmd.testpackage: pkg.install
	$(MAKE) -C cmd testpackage

%.clean:
	$(MAKE) -C $* clean

%.nuke:
	$(MAKE) -C $* nuke

pkg.test:
	$(MAKE) -C pkg test

cmd.test: pkg.install
	$(MAKE) -C cmd test

%.bench:
	$(MAKE) -C $* bench

gofmt:
	@find -name '*.go' | xargs gofmt -w=true -s=true

govet:
	@govet .

.PHONY: gofmt govet
