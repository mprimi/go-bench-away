package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/client"
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

func (cmd *initCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
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
		client.CreateJobsQueue,
		client.CreateJobsRepository,
		client.CreateArtifactsStore,
	}

	for _, fun := range initFuncs {
		if err := fun(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
