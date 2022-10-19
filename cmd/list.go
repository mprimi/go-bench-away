package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

type listCmd struct {
	baseCommand
}

func listCommand() subcommands.Command {
	return &listCmd{
		baseCommand: baseCommand{
			name:     "list",
			synopsis: "lists recent jobs",
			usage:    "list [options]",
			setFlags: func(_ *flag.FlagSet) {},
		},
	}
}

func (cmd *listCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	nc, js, connErr := core.Connect(rootOpts.natsServerUrl, "go-bench-away CLI")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "%v\n", connErr)
		return subcommands.ExitFailure
	}
	defer nc.Close()

	err := core.List(js, 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
