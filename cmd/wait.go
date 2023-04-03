package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mprimi/go-bench-away/pkg/client"
	"github.com/mprimi/go-bench-away/pkg/core"

	"github.com/google/subcommands"
)

type waitCmd struct {
	baseCommand
}

func waitCommand() subcommands.Command {
	return &waitCmd{
		baseCommand: baseCommand{
			name:     "wait",
			synopsis: "Waits for a set of jobs to complete",
			usage:    "wait <jobId> [jobId [...]]\n",
		},
	}
}

func (cmd *waitCmd) SetFlags(f *flag.FlagSet) {
}

func (cmd *waitCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Missing job ID arguments\n")
		return subcommands.ExitUsageError
	}

	jobIds := f.Args()

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.InitJobsRepository(),
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer c.Close()

	wg := sync.WaitGroup{}

	const kStatusPollInterval = 3 * time.Second

	waitJob := func(jobId string) {
		defer wg.Done()

		fmt.Printf("Waiting for job %s\n", jobId)

		previousRevision := uint64(0)

		for {
			job, revision, err := c.LoadJob(jobId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to wait on job %s: %v\n", jobId, err)
				return
			}

			if revision != previousRevision {
				fmt.Printf("%s: %s\n", jobId, job.Status)
			}

			if job.Status == core.Failed || job.Status == core.Succeeded {
				return
			}

			previousRevision = revision
			time.Sleep(kStatusPollInterval)
		}
	}

	for _, jobId := range jobIds {
		wg.Add(1)
		go waitJob(jobId)
	}

	wg.Wait()

	return subcommands.ExitSuccess
}
