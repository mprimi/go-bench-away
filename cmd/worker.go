package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/internal/worker"
	"github.com/mprimi/go-bench-away/pkg/client"

	"github.com/google/subcommands"
)

type workerCmd struct {
	baseCommand
	jobsDir string
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

func (cmd *workerCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.jobsDir, "jobs_dir", "", "Directory where jobs are staged (defaults to os.MkdirTemp)")
}

func (cmd *workerCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.Verbose(rootOptions.verbose),
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.WithClientName("go-bench-away Worker"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer c.Close()

	if cmd.jobsDir != "" {
		err := os.MkdirAll(cmd.jobsDir, 0750)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating jobs directory: %v\n", err)
			return subcommands.ExitFailure
		}
	}

	w, err := worker.NewWorker(c, cmd.jobsDir)
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
