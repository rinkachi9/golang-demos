# Installation

* `sudo apt-get update`
* `sudo apt-get install -y clang llvm`

Command: `go run github.com/cilium/ebpf/cmd/bpf2go -target native -go-package main Counter bpf/counter.c`

Run: `go run .`

Warning: it requires sudo permission.
