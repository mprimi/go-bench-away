package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/internal/worker"
	"github.com/mprimi/go-bench-away/v1/client"

	"github.com/google/subcommands"
)

type workerCmd struct {
	baseCommand
	jobsDir             string
	altQueue            string
	gitRemoteFilterExpr string
}

func workerCommand() subcommands.Command {
	return &workerCmd{
		baseCommand: baseCommand{
			name:     "worker",
			synopsis: "starts a benchmark worker",
			usage:    "worker [options]\n",
		},
	}
}

func (cmd *workerCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.jobsDir, "jobs_dir", "", "Directory where jobs are staged (defaults to os.MkdirTemp)")
	f.StringVar(&cmd.altQueue, "queue", "", "Consume job from a non-default queue with the specified name")
	f.StringVar(&cmd.gitRemoteFilterExpr, "gitRemoteFilterExpr", "", "Regex to restrict which git remotes can be targeted")
}

func (cmd *workerCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	clientOpts := []client.Option{
		client.Verbose(rootOptions.verbose),
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.WithClientName("go-bench-away Worker"),
	}

	if cmd.altQueue != "" {
		clientOpts = append(
			clientOpts,
			client.WithAltQueue(cmd.altQueue),
			client.WithClientName(fmt.Sprintf("go-bench-away Worker (queue: %s", cmd.altQueue)),
		)
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

	if cmd.jobsDir != "" {
		err := os.MkdirAll(cmd.jobsDir, 0750)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating jobs directory: %v\n", err)
			return subcommands.ExitFailure
		}
	}

	var allowedGitRemoteExpr []string
	if cmd.gitRemoteFilterExpr != "" {
		allowedGitRemoteExpr = []string{
			cmd.gitRemoteFilterExpr,
		}
	}

	w, err := worker.NewWorker(c, cmd.jobsDir, allowedGitRemoteExpr)
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
