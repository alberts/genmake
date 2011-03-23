include Make.dirs

all: make

make: $(addsuffix .make, $(DIRS))
testpackage: $(addsuffix .testpackage, $(DIRS))
clean: $(addsuffix .clean, $(DIRS))
nuke: $(addsuffix .nuke, $(DIRS))
install: make
test: make $(addsuffix .test, $(TEST))
bench: make $(addsuffix .bench, $(BENCH))

%.make:
	$(MAKE) -C $*

%.testpackage: %.make
	$(MAKE) -C $* testpackage

%.install:
	$(MAKE) -C $* install

%.clean:
	$(MAKE) -C $* clean

%.nuke:
	$(MAKE) -C $* nuke

%.test: %.make
	$(MAKE) -C $* test

%.bench:
	$(MAKE) -C $* bench
