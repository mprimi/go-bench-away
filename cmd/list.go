package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mprimi/go-bench-away/v1/client"
	"github.com/mprimi/go-bench-away/v1/core"

	"github.com/google/subcommands"
)

type listCmd struct {
	baseCommand
	limit    int
	altQueue string
}

func listCommand() subcommands.Command {
	return &listCmd{
		baseCommand: baseCommand{
			name:     "list",
			synopsis: "lists recent jobs",
			usage:    "list [options]\n",
		},
	}
}

func (cmd *listCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.limit, "n", 10, "Maximum number of recent jobs to show (0 for unlimited)")
	f.StringVar(&cmd.altQueue, "queue", "", "Read jobs from a non-default queue with the specified name")
}

func (cmd *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	clientOpts := []client.Option{
		client.Verbose(rootOptions.verbose),
		client.InitJobsQueue(),
		client.InitJobsRepository(),
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

	jobs, err := c.LoadRecentJobs(cmd.limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	if len(jobs) == 0 {
		fmt.Printf("No jobs found\n")
		return subcommands.ExitSuccess
	}

	fmt.Printf("Recent jobs:\n")
	for _, job := range jobs {

		fmt.Printf(
			" %s %s [%v]\n"+
				"     - Submitted: %v (%v ago) by %s\n"+
				"     - Remote: %s Ref: %s\n"+
				"     - Filter: '%s'\n"+
				"     - Repetitions: %d x %v\n"+
				"",
			job.Status.Icon(),
			job.Id,
			job.Status,
			job.Created,
			time.Since(job.Created).Truncate(time.Minute),
			job.Parameters.Username,
			job.Parameters.GitRemote,
			job.Parameters.GitRef,
			job.Parameters.TestsFilterExpr,
			job.Parameters.Reps,
			job.Parameters.TestMinRuntime,
		)

		switch job.Status {
		case core.Failed:
			fallthrough
		case core.Succeeded:
			fmt.Printf(
				"     - Run time: %v\n"+
					"     - SHA: %s\n"+
					"     - Go: %s\n"+
					"     - Log file: %s\n"+
					"     - Results file: %s\n"+
					"",
				job.RunTime(),
				job.SHA,
				job.GoVersion,
				job.Log,
				job.Results,
			)

		case core.Running:
			fmt.Printf(
				"     - Run time: %v (max: %v)\n"+
					"",
				job.RunTime(),
				job.Parameters.Timeout,
			)

		case core.Submitted:
			//NOOP
		}
		fmt.Printf("\n")
	}

	return subcommands.ExitSuccess
}
