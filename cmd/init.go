package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/v1/client"

	"github.com/google/subcommands"
)

type initCmd struct {
	baseCommand
	altQueue string
}

func initCommand() subcommands.Command {
	return &initCmd{
		baseCommand: baseCommand{
			name:     "init",
			synopsis: "Initializes server schemas (Stream, KV store, Object store)",
			usage:    "init [options]\n",
		},
	}
}

func (cmd *initCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.altQueue, "queue", "", "Initialize non-default jobs queue with the given name")
}

func (cmd *initCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
		c.CreateJobsQueue,
		c.CreateJobsRepository,
		c.CreateArtifactsStore,
	}

	for _, fun := range initFuncs {
		if err := fun(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
