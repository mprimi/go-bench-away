package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/v1/client"
)

type cancelCmd struct {
	baseCommand
}

func cancelCommand() subcommands.Command {
	return &cancelCmd{
		baseCommand: baseCommand{
			name:     "cancel",
			synopsis: "Cancel a queued job",
			usage:    "cancel [options] jobId [jobId [...]]\n",
		},
	}
}

func (cmd *cancelCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *cancelCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer c.Close()

	for _, jobId := range f.Args() {
		err = c.CancelJob(jobId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return subcommands.ExitFailure
		}
		fmt.Printf("Cancelled job: %s\n", jobId)
	}

	return subcommands.ExitSuccess
}
