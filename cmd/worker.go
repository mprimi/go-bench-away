package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

type workerCmd struct {
	baseCommand
}

func workerCommand() subcommands.Command {
	return &workerCmd{
		baseCommand: baseCommand{
			name:     "worker",
			synopsis: "starts a benchmark worker",
			usage:    "worker [options]",
			setFlags: func(_ *flag.FlagSet) {},
		},
	}
}

func (ec *workerCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	nc, js, connErr := core.Connect(rootOpts.natsServerUrl, "go-bench-away Worker")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", connErr)
		return subcommands.ExitFailure
	}
	defer nc.Close()

	err := core.RunWorker(js)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
