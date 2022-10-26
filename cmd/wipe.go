package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/client"
)

type wipeCmd struct {
	baseCommand
}

func wipeCommand() subcommands.Command {
	return &wipeCmd{
		baseCommand: baseCommand{
			name:     "wipe",
			synopsis: "Deletes server schemas (Stream, KV store, Object store)",
			usage:    "wipe\n",
		},
	}
}

func (cmd *wipeCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	client, err := client.NewClient(
		rootOpts.natsServerUrl,
		rootOpts.credentials,
		rootOpts.namespace,
		client.Verbose(rootOpts.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer client.Close()

	initFuncs := []func() error{
		client.DeleteJobsQueue,
		client.DeleteJobsRepository,
		client.DeleteArtifactsStore,
	}

	for _, fun := range initFuncs {
		if err := fun(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
