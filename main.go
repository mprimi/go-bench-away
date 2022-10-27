package main

import (
	"os"

	"github.com/mprimi/go-bench-away/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
