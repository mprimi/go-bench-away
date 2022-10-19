package main

import (
	"os"

	"github.com/mprimi/go-bench-away/cmd"
)

const (
	name = "go-bench-away"
)

// Substituted at build time by goreleaser
var version = "dev"
var sha = "?"
var buildDate = "?"

func main() {
	os.Exit(cmd.Run(name, version, sha, buildDate, os.Args[1:]))
}
