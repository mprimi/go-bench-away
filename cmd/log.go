package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mprimi/go-bench-away/v1/client"

	"github.com/google/subcommands"
)

type logCmd struct {
	baseCommand
}

func logCommand() subcommands.Command {
	return &logCmd{
		baseCommand: baseCommand{
			name:     "log",
			synopsis: "Shows the log file of a completed job",
			usage:    "log <jobId>\n",
		},
	}
}

func (cmd *logCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *logCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Must pass one job ID argument\n")
		return subcommands.ExitUsageError
	}

	jobId := f.Args()[0]

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer c.Close()

	job, _, err := c.LoadJob(jobId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	if job.Log == "" {
		fmt.Printf("No log artifact for job %s\n", job.Id)
	} else {
		err := c.LoadLogArtifact(job, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
			return subcommands.ExitFailure
		}

		fmt.Fprintf(os.Stdout, "\n")
	}

	return subcommands.ExitSuccess
}
