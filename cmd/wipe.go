package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

type wipeCmd struct {
	baseCommand
}

func wipeCommand() subcommands.Command {
	return &wipeCmd{
		baseCommand: baseCommand{
			name:     "init",
			synopsis: "Initializes server schemas (Stream, KV store, Object store)",
			usage:    "init",
		},
	}
}

func (cmd *wipeCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	nc, js, connErr := core.Connect(rootOpts.natsServerUrl, "go-bench-away CLI")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", connErr)
		return subcommands.ExitFailure
	}
	defer nc.Close()

	err := core.WipeSchema(js)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v", cmd.name, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
