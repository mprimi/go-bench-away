package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/worker"
)

type workerCmd struct {
	baseCommand
}

func workerCommand() subcommands.Command {
	return &workerCmd{
		baseCommand: baseCommand{
			name:     "worker",
			synopsis: "starts a benchmark worker",
			usage:    "worker [options]\n",
			setFlags: func(_ *flag.FlagSet) {},
		},
	}
}

func (cmd *workerCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	client, err := client.NewClient(
		rootOpts.natsServerUrl,
		rootOpts.credentials,
		rootOpts.namespace,
		client.Verbose(rootOpts.verbose),
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.WithClientName("go-bench-away Worker"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer client.Close()

	w, err := worker.NewWorker(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	err = w.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
