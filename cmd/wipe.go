package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/pkg/client"

	"github.com/google/subcommands"
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

func (cmd *wipeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer c.Close()

	initFuncs := []func() error{
		c.DeleteJobsQueue,
		c.DeleteJobsRepository,
		c.DeleteArtifactsStore,
	}

	for _, fun := range initFuncs {
		if err := fun(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
