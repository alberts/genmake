package main

import (
	"strings"
)

// TODO bsd => freebsd || darwin
// TODO unix => linux || freebsd || darwin

func validGOOS(os string) bool {
	return os == "linux"
}

func validGOARCH(arch string) bool {
	return arch == "amd64" || arch == "386"
}

func goosArchSource(name string) (os, arch string, ok bool) {
	parts := strings.Split(name, "_", -1)
	if len(parts) < 2 {
		return "", "", false
	}
	os = parts[len(parts)-2]
	arch = strings.Split(parts[len(parts)-1], ".", 2)[0]
	if validGOOS(os) && validGOARCH(arch) {
		return os, arch, true
	}
	return "", "", false
}

func goarchSource(name string) (string, bool) {
	parts := strings.Split(name, "_", -1)
	if len(parts) <= 1 {
		return "", false
	}
	os := strings.Split(parts[len(parts)-1], ".", 2)[0]
	if validGOARCH(os) {
		return os, true
	}
	return "", false
}

func goosSource(name string) (string, bool) {
	parts := strings.Split(name, "_", -1)
	if len(parts) <= 1 {
		return "", false
	}
	os := strings.Split(parts[len(parts)-1], ".", 2)[0]
	if validGOOS(os) {
		return os, true
	}
	return "", false
}
