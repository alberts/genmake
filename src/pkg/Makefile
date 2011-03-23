include $(GOROOT)/src/Make.inc

include Make.dirs

all: make

make: $(addsuffix .install, $(DIRS))
testpackage: $(addsuffix .testpackage, $(DIRS))
clean: nuke
nuke: $(addsuffix .nuke, $(DIRS))
install: make
test: make $(addsuffix .test, $(TEST))
bench: make $(addsuffix .bench, $(BENCH))

%.install:
	$(MAKE) -C $* install

%.testpackage: %.install
	$(MAKE) -C $* testpackage

%.nuke:
	$(MAKE) -C $* nuke

%.test: %.install
	$(MAKE) -C $* test

%.bench:
	$(MAKE) -C $* bench

include Make.deps
