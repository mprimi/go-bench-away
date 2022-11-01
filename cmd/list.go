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

type listCmd struct {
	baseCommand
	limit int
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
}

func (cmd *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	client, err := client.NewClient(
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
	defer client.Close()

	jobs, err := client.LoadRecentJobs(cmd.limit)
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
				"     - Submitted: %v ago by %s\n"+
				"     - Remote: %s Ref: %s\n"+
				"     - Filter: '%s'\n"+
				"",
			job.Status.Icon(),
			job.Id,
			job.Status,
			time.Since(job.Created).Truncate(time.Minute),
			job.Parameters.Username,
			job.Parameters.GitRemote,
			job.Parameters.GitRef,
			job.Parameters.TestsFilterExpr,
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
				job.Completed.Sub(job.Started).Round(time.Second),
				job.SHA,
				job.GoVersion,
				job.Log,
				job.Results,
			)

		case core.Running:
			fmt.Printf(
				"     - Run time: %v\n"+
					"",
				time.Since(job.Started).Round(time.Second),
			)

		case core.Submitted:
			//NOOP
		}
		fmt.Printf("\n")
	}

	return subcommands.ExitSuccess
}
