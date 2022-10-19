package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

type waitCmd struct {
	baseCommand
}

func waitCommand() subcommands.Command {
	return &waitCmd{
		baseCommand: baseCommand{
			name:     "wait",
			synopsis: "Waits for a given job completion",
			usage:    "wait <jobId>",
		},
	}
}

func (cmd *waitCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *waitCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Missing job ID argument\n")
		return subcommands.ExitUsageError
	}

	nc, js, connErr := core.Connect(rootOpts.natsServerUrl, "go-bench-away CLI")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", connErr)
		return subcommands.ExitFailure
	}
	defer nc.Close()

	err := core.WaitJob(js, f.Args()[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v", cmd.name, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
