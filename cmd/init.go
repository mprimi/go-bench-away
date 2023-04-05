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
}

func initCommand() subcommands.Command {
	return &initCmd{
		baseCommand: baseCommand{
			name:     "init",
			synopsis: "Initializes server schemas (Stream, KV store, Object store)",
			usage:    "init\n",
		},
	}
}

func (cmd *initCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
