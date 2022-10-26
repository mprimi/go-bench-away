package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/client"
	"github.com/mprimi/go-bench-away/internal/core"
)

type waitCmd struct {
	baseCommand
}

func waitCommand() subcommands.Command {
	return &waitCmd{
		baseCommand: baseCommand{
			name:     "wait",
			synopsis: "Waits for a given job completion",
			usage:    "wait <jobId>\n",
		},
	}
}

func (cmd *waitCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *waitCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Missing job ID argument\n")
		return subcommands.ExitUsageError
	}

	jobId := f.Args()[0]

	client, err := client.NewClient(
		rootOpts.natsServerUrl,
		rootOpts.credentials,
		rootOpts.namespace,
		client.InitJobsRepository(),
		client.Verbose(rootOpts.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer client.Close()

	fmt.Printf("Waiting for job %s\n", jobId)

	previousRevision := uint64(0)

pollLoop:
	for {
		job, revision, err := client.LoadJob(jobId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return subcommands.ExitFailure
		}

		if revision != previousRevision {
			fmt.Printf("%s\n", job.Status)
		}

		if job.Status == core.Failed || job.Status == core.Succeeded {
			break pollLoop
		}

		previousRevision = revision
		time.Sleep(1 * time.Second)
	}

	return subcommands.ExitSuccess
}
