package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/v1/client"

	"github.com/google/subcommands"
)

type wipeCmd struct {
	baseCommand
	altQueue string
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

func (cmd *wipeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.altQueue, "queue", "", "Wipe non-default queue with the specified name")
}

func (cmd *wipeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	clientOpts := []client.Option{
		client.Verbose(rootOptions.verbose),
	}

	if cmd.altQueue != "" {
		clientOpts = append(clientOpts, client.WithAltQueue(cmd.altQueue))
	}

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		clientOpts...,
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
