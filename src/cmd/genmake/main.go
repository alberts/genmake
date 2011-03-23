package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const openFlags int = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

func fprintf(file *os.File, format string, a ...interface{}) {
	if _, err := fmt.Fprintf(file, format, a...); err != nil {
		panic(err)
	}
}

type filesMap map[string]sort.StringArray

func (fm filesMap) append(key string, elem string) {
	if _, ok := fm[key]; !ok {
		fm[key] = make([]string, 0)
	}
	fm[key] = append(fm[key], elem)
	sort.Sort(fm[key])
}

type target struct {
	path      string
	imports   map[string]bool
	makeinc   bool
	gofiles   sort.StringArray
	cgo       sort.StringArray
	pb        sort.StringArray
	gotest    sort.StringArray
	goarch    filesMap
	goos      filesMap
	goosArch  filesMap
	cgoarch   filesMap
	cgoos     filesMap
	cgoosArch filesMap
}

func (this *target) empty() bool {
	if len(this.gofiles) > 0 || len(this.cgo) > 0 || len(this.pb) > 0 {
		return false
	}
	if len(this.goosArch) > 0 || len(this.goarch) > 0 || len(this.goos) > 0 {
		return false
	}
	if len(this.cgoosArch) > 0 || len(this.cgoarch) > 0 || len(this.cgoos) > 0 {
		return false
	}
	return true
}

func appendFiles(files map[string]bool, more filesMap) {
	for _, v := range more {
		for _, filename := range v {
			files[filename] = true
		}
	}
}

func (this *target) deps(targets targetsMap) (internal, external sort.StringArray) {
	for imp := range this.imports {
		if _, ok := targets[imp]; ok {
			internal = append(internal, imp)
		} else {
			external = append(external, imp)
		}
	}
	sort.Sort(internal)
	sort.Sort(external)
	return
}

func (this *target) write(file *os.File) {
	gofiles := this.gofiles
	for _, proto := range this.pb {
		// strip .proto suffix and add .pb.go
		gofiles = append(gofiles, proto[0:len(proto)-6]+".pb.go")
	}
	sort.Sort(gofiles)
	writeFiles(file, "GOFILES", gofiles, this.goarch, this.goos, this.goosArch)
	writeFiles(file, "GOTESTFILES", this.gotest, nil, nil, nil)
	writeFiles(file, "CGOFILES", this.cgo, this.cgoarch, this.cgoos, this.cgoosArch)
}

func writeOsArch(file *os.File, basevar string, more []string, files filesMap) {
	if len(files) == 0 {
		return
	}
	var morevar string
	if len(more) == 1 {
		morevar = "$(" + more[0] + ")"
	} else {
		morevar = "$(" + more[0] + ")_$(" + more[1] + ")"
	}
	fprintf(file, "%s+=$(%s_%s)\n\n", basevar, basevar, morevar)
	for avar, afiles := range files {
		fprintf(file, "%s_%s=\\\n", basevar, avar)
		for _, afile := range afiles {
			fprintf(file, "\t%s\\\n", afile)
		}
		fprintf(file, "\n")
	}
}

func writeFiles(file *os.File, basevar string, files []string, goarch, goos, goosArch filesMap) {
	if len(files) == 0 && len(goarch) == 0 && len(goos) == 0 && len(goosArch) == 0 {
		return
	}
	if len(files) > 0 {
		fprintf(file, "%s=\\\n", basevar)
		for _, name := range files {
			fprintf(file, "\t%s\\\n", name)
		}
		fprintf(file, "\n")
	}
	writeOsArch(file, basevar, []string{"GOARCH"}, goarch)
	writeOsArch(file, basevar, []string{"GOOS"}, goos)
	writeOsArch(file, basevar, []string{"GOOS", "GOARCH"}, goosArch)
}

func newTarget() *target {
	return &target{"", make(map[string]bool), false, nil, nil, nil, nil,
		make(filesMap), make(filesMap), make(filesMap), make(filesMap), make(filesMap), make(filesMap)}
}

type targetsMap map[string]*target

type srcWalker struct {
	srcDir  string
	pkg     bool
	targets targetsMap
}

func (this *srcWalker) finish() targetsMap {
	for k, target := range this.targets {
		if target.empty() {
			this.targets[k] = nil, false
		}
	}
	return this.targets
}

func (this *srcWalker) VisitDir(name string, f *os.FileInfo) bool {
	base := filepath.Base(name)
	if base == "_obj" || base == "_test" || base == "testdata" {
		return false
	}
	return true
}

type importVisitor struct {
	target *target
	cgo    bool
}

func (v *importVisitor) Visit(node ast.Node) ast.Visitor {
	spec, ok := node.(*ast.ImportSpec)
	if !ok {
		return v
	}
	path := string(spec.Path.Value)
	path = path[1 : len(path)-1]
	if path == "C" {
		v.cgo = true
	}
	if path != "C" && path != "unsafe" {
		v.target.imports[path] = true
	}
	return v
}

func (this *srcWalker) VisitFile(name string, f *os.FileInfo) {
	dir, _ := filepath.Split(name)
	dir = filepath.Clean(dir)
	if dir == this.srcDir {
		return
	}

	base := filepath.Base(name)
	if strings.HasPrefix(base, "_cgo") {
		return
	}
	if strings.HasSuffix(base, ".pb.go") {
		return
	}
	if base == "_testmain.go" {
		return
	}

	var targ string
	if this.pkg {
		targ = name[len(this.srcDir)+1 : len(name)-len(base)-1]
	} else {
		_, targ = filepath.Split(dir)
	}

	if this.targets == nil {
		this.targets = make(targetsMap)
	}
	if _, ok := this.targets[targ]; !ok {
		this.targets[targ] = newTarget()
	}
	target := this.targets[targ]
	target.path = dir

	if base == "Make.inc" {
		target.makeinc = true
		return
	}

	if strings.HasSuffix(base, ".proto") {
		target.pb = append(target.pb, base)
		sort.Sort(target.pb)
		return
	}

	if !strings.HasSuffix(base, ".go") {
		return
	}

	if strings.HasSuffix(base, "_test.go") {
		target.gotest = append(target.gotest, base)
		sort.Sort(target.gotest)
		return
	}

	fileNode, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
	if err != nil {
		panic(name + ": " + err.String())
	}
	v := &importVisitor{target, false}
	ast.Walk(v, fileNode)
	cgo := v.cgo

	if os, arch, ok := goosArchSource(base); ok {
		key := os + "_" + arch
		if cgo && this.pkg {
			target.cgoosArch.append(key, base)
		} else {
			target.goosArch.append(key, base)
		}
	} else if arch, ok := goarchSource(base); ok {
		if cgo && this.pkg {
			target.cgoarch.append(arch, base)
		} else {
			target.goarch.append(arch, base)
		}
	} else if os, ok := goosSource(base); ok {
		if cgo && this.pkg {
			target.cgoos.append(os, base)
		} else {
			target.goos.append(os, base)
		}
	} else {
		if cgo && this.pkg {
			target.cgo = append(target.cgo, base)
			sort.Sort(target.cgo)
		} else {
			target.gofiles = append(target.gofiles, base)
			sort.Sort(target.gofiles)
		}
	}
}

func printDeps(file *os.File, targ string, target *target, targets targetsMap, cmd bool) {
	internal, _ := target.deps(targets)
	if len(internal) == 0 {
		return
	}

	var up string
	if cmd {
		up = "../../pkg/"
	} else {
		nlevels := len(strings.Split(targ, "/", -1))
		for i := 0; i < nlevels; i++ {
			up += "../"
		}
	}

	fprintf(file, "PREREQ+=\\\n")
	for _, dep := range internal {
		fprintf(file, "\t$(QUOTED_GOROOT)/pkg/$(GOOS)_$(GOARCH)/%s.a\\\n", dep)
	}
	fprintf(file, "\n")
	for _, dep := range internal {
		fprintf(file, "$(QUOTED_GOROOT)/pkg/$(GOOS)_$(GOARCH)/%s.a:\n", dep)
		fprintf(file, "\t$(MAKE) -C %s%s install\n\n", up, dep)
	}
	fprintf(file, ".DEFAULT_GOAL:=\n\n")
}

func makePkg(pkg string, target *target, targets targetsMap) {
	makefile := filepath.Join(target.path, "Makefile")
	file, err := os.Open(makefile, openFlags, 0666)
	if err != nil {
		panic(err)
	}

	fprintf(file, "include $(GOROOT)/src/Make.inc\n\n")
	fprintf(file, "TARG=%s\n\n", pkg)

	target.write(file)

	if len(target.pb) > 0 {
		fprintf(file, "CLEANFILES+=*.pb.go\n\n")
		fprintf(file, "%%.pb.go: *.proto\n\tprotoc --go_out=. *.proto\n\n")
	}

	if target.makeinc {
		fprintf(file, "include Make.inc\n\n")
	}

	printDeps(file, pkg, target, targets, false)

	fprintf(file, "include $(GOROOT)/src/Make.pkg\n")

	if err := file.Close(); err != nil {
		panic(err)
	}
}

func makeCmd(cmd string, target *target, targets targetsMap) {
	makefile := filepath.Join(target.path, "Makefile")
	file, err := os.Open(makefile, openFlags, 0666)
	if err != nil {
		panic(err)
	}

	fprintf(file, "include $(GOROOT)/src/Make.inc\n\n")
	fprintf(file, "TARG=%s\n\n", cmd)
	target.write(file)

	if target.makeinc {
		fprintf(file, "include Make.inc\n\n")
	}

	printDeps(file, cmd, target, targets, true)

	fprintf(file, "include $(GOROOT)/src/Make.cmd\n")

	if err := file.Close(); err != nil {
		panic(err)
	}
}

func makeDirs(dir string, targets targetsMap) {
	var dirs sort.StringArray
	var notest sort.StringArray
	for _, target := range targets {
		targ := target.path[len(dir)+1:]
		dirs = append(dirs, targ)
		if len(target.gotest) == 0 {
			notest = append(notest, targ)
		}
	}
	sort.Sort(dirs)
	sort.Sort(notest)

	file, err := os.Open(filepath.Join(dir, "Make.dirs"), openFlags, 0666)
	if err != nil {
		panic(err)
	}
	fprintf(file, "DIRS=\\\n")
	for _, dir := range dirs {
		fprintf(file, "\t%s\\\n", dir)
	}

	fprintf(file, "\n")

	fprintf(file, "NOTEST=\\\n")
	for _, dir := range notest {
		fprintf(file, "\t%s\\\n", dir)
	}

	fprintf(file, "\n")

	fprintf(file, "TEST=$(filter-out $(NOTEST),$(DIRS))\n\n")
	fprintf(file, "BENCH=$(filter-out $(NOBENCH),$(TEST))\n")

	if err := file.Close(); err != nil {
		panic(err)
	}
}

func makeGitIgnore(dir string, targets targetsMap) {
	var targs sort.StringArray
	for _, target := range targets {
		targ := target.path[len(dir)+1:]
		targs = append(targs, targ)
	}
	sort.Sort(targs)
	file, err := os.Open(filepath.Join(dir, ".gitignore"), openFlags, 0666)
	if err != nil {
		panic(err)
	}
	for _, targ := range targs {
		fprintf(file, "%s/%s\n", targ, targ)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}

func makeDeps(dir string, targets targetsMap) {
	file, err := os.Open(filepath.Join(dir, "Make.deps"), openFlags, 0666)
	if err != nil {
		panic(err)
	}
	for pkg, target := range targets {
		internal, external := target.deps(targets)
		fprintf(file, "%s.install:", pkg)
		for _, imp := range internal {
			fprintf(file, " %s.install", imp)
		}
		for _, imp := range external {
			// skip external dependencies for now
			continue
			fprintf(file, " $(GOROOT)/pkg/$(GOOS)_$(GOARCH)/%s.a", imp)
		}
		fprintf(file, "\n")
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pkgDir := filepath.Clean(filepath.Join(wd, "src", "pkg"))
	pkgWalker := &srcWalker{srcDir: pkgDir, pkg: true}
	filepath.Walk(pkgDir, pkgWalker, nil)
	pkgTargets := pkgWalker.finish()
	for pkg, target := range pkgTargets {
		makePkg(pkg, target, pkgTargets)
	}
	makeDirs(pkgDir, pkgTargets)
	makeDeps(pkgDir, pkgTargets)

	cmdDir := filepath.Clean(filepath.Join(wd, "src", "cmd"))
	cmdWalker := &srcWalker{srcDir: cmdDir, pkg: false}
	filepath.Walk(cmdDir, cmdWalker, nil)
	cmdTargets := cmdWalker.finish()
	for cmd, target := range cmdTargets {
		makeCmd(cmd, target, pkgTargets)
	}
	makeDirs(cmdDir, cmdTargets)
	makeGitIgnore(cmdDir, cmdTargets)
}
